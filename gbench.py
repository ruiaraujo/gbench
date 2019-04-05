import os
import re

import asyncio
import socket
import time
from urllib.parse import urlparse
from asyncio.subprocess import PIPE
import json
import yaml
from datetime import timedelta
import fire

dir_path = os.path.dirname(os.path.realpath(__file__))

regex = re.compile(r'((?P<hours>\d+?)hr)?((?P<minutes>\d+?)m)?((?P<seconds>\d+?)s)?')


def ensure_folder(dir_name):
    if not os.path.exists(dir_name):
        os.makedirs(dir_name)


def parse_duration(time_str):
    if not time_str:
        return 0
    parts = regex.match(str(time_str))
    if not parts:
        return
    parts = parts.groupdict()
    time_params = {}
    for (name, param) in parts.items():
        if param:
            time_params[name] = int(param)
    return timedelta(**time_params).total_seconds()


async def run_wrk(endpoint, query, concurrency, duration, rate):
    wrk = await asyncio.create_subprocess_exec(
        'wrk2', '-t', '1', '-R', str(rate), '--latency', '-c', str(concurrency), '-d', str(int(duration)), '-s',
        'misc/graphql.lua', endpoint, query, stdout=PIPE, stderr=PIPE)

    output, json_data = await wrk.communicate()
    log = output.decode('utf-8')
    first_del = "Detailed Percentile spectrum:"
    second_del = "----------------------------------------------------------"
    hdr_histogram = log[log.find(first_del) + len(first_del):log.find(second_del)]
    print(log)
    try:
        data = json.loads(json_data)
    except:
        raise Exception("Received invalid json from wrk: {}".format(json_data))

    response_json = data.pop('response', None)
    if response_json:
        try:
            response = json.loads(response_json)
        except:
            raise Exception("Received invalid json from the GraphQL query: {}".format(response_json))
    else:
        response = None
    await wrk.wait()

    return hdr_histogram, response


async def start_server(command, cwd):
    server = await asyncio.create_subprocess_exec(*command.split(), cwd=cwd)
    return server


async def get_query_result(query_name, server_name, url, query_filename, rate, duration, concurrency,
                           warmup_duration, warmup_concurrency):
    print("- Running query {}".format(query_name, server_name))
    if warmup_duration:
        print("  > Warming up the server")
        await run_wrk(url, query_filename, duration=warmup_duration, rate=rate, concurrency=warmup_concurrency)
        await asyncio.sleep(2)
    print("  > Benchmarking")
    histogram, response = await run_wrk(url, query_filename, duration=duration, rate=rate, concurrency=concurrency)
    return ({
                "query_name": query_name,
                "server_name": server_name,
                "histogram": histogram,
            }, response)


async def force_termination(pid):
    kill = await asyncio.create_subprocess_exec(
        'kill', '-9', str(pid))
    await kill.communicate()


def wait_for_port(port, host='localhost', timeout=60):
    """Wait until a port starts accepting TCP connections.
    Args:
        port (int): Port number.
        host (str): Host address on which the port should exist.
        timeout (float): In seconds. How long to wait before raising errors.
    Raises:
        TimeoutError: The port isn't accepting connection after time specified in `timeout`.
    """
    start_time = time.perf_counter()
    print("Waiting {}s for {}:{}".format(timeout, host, port))
    while True:
        try:
            with socket.create_connection((host, port), timeout=timeout):
                break
        except OSError as ex:
            time.sleep(0.01)
            if time.perf_counter() - start_time >= timeout:
                raise TimeoutError('Waited too long for the port {} on host {} to start accepting '
                                   'connections.'.format(port, host)) from ex


async def bench_server(name, command, cwd, output_folder, endpoint, queries, warmup_duration,
                       warmup_concurrency, rate, duration, concurrency):
    print("Starting server: {}".format(name))
    server = await start_server(command, cwd)
    wait_for_port(urlparse(endpoint).port)
    print("Detected server start")
    try:
        for query in queries:
            result, response = None, None
            try:
                result, response = await get_query_result(query.get('name'), name, endpoint, query.get('filename'),
                                                          rate, duration, concurrency, warmup_duration, warmup_concurrency)
                dir_name = os.path.join(output_folder, query.get('name').lower())
                ensure_folder(dir_name)
                with open(os.path.join(dir_name, name.lower()), "w") as output:
                    output.write(result.get("histogram"))
                    output.close()
                # We wait 2 seconds between trials
                await asyncio.sleep(2)
            except Exception as e:
                print("- Exception occured: {} - SKIPPING".format(str(e)))

            expected_result = query.get('expectedResult')
            if expected_result and expected_result != response:
                raise Exception("GraphQL response mismatch expected.\nReceived: {}\nExpected:{}".format(
                    result['response'], expected_result
                ))

    except TimeoutError as te:
        print(te)
    except:
        raise
    finally:
        print("Terminating server: {}".format(name))
        if server.returncode is None:
            server.terminate()
        await server.wait()
        await asyncio.sleep(1)
        print("Terminated")


class Program(object):
    """The GraphQL Benchmarking utility"""
    @staticmethod
    def benchmark(config, output_folder='./results'):
        """Starts a benchmark given config and an optional output file"""
        loop = asyncio.get_event_loop()
        loop.run_until_complete(_main(config, output_folder))


async def _main(config, output_folder):
    with open(config, 'r') as content_file:
        content = content_file.read()
        config = yaml.load(content)

        load_config = config.get('load')
        all_queries = config.get('queries')
        all_servers = config.get('servers')
        for query in all_queries:
            filename = query.get('expectedResultFilename')
            if filename:
                with open(filename, 'r') as file:
                    content = file.read()
                    query['expectedResult'] = json.loads(content)
        load_duration = parse_duration(load_config.get('duration') or 120)
        load_rate = load_config.get('rate') or 500
        load_concurrency = load_config.get('concurrency') or 20

        ensure_folder(output_folder)
        for server in all_servers:
            server_name = server.get('name')
            server_run = server.get('run')
            server_command = server_run.get('command')
            server_cwd = server_run.get('cwd') or "./"
            server_endpoint = server.get('endpoint')
            server_warmup = server.get('warmup') or {}
            server_warmup_duration = parse_duration(server_warmup.get('duration') or 0)
            server_warmup_concurrency = server_warmup.get('concurrency') or 1
            await bench_server(
                name=server_name,
                command=server_command,
                output_folder=output_folder,
                cwd=os.path.join(dir_path, server_cwd),
                endpoint=server_endpoint,
                queries=all_queries,
                warmup_duration=server_warmup_duration,
                warmup_concurrency=server_warmup_concurrency,
                concurrency=load_concurrency,
                duration=load_duration,
                rate=load_rate,
            )




if __name__ == "__main__":
    fire.Fire(Program)
