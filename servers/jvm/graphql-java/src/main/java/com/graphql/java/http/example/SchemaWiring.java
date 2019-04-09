package com.graphql.java.http.example;

import graphql.schema.DataFetcher;
import graphql.schema.GraphQLObjectType;
import graphql.schema.TypeResolver;

import java.util.ArrayList;
import java.util.List;

/**
 * This is our wiring used to put fetching behaviour behind a graphql field.
 */
public class SchemaWiring {

    static List<String> listOfString = generateStringList();
    static List<Object> listOfObject = generateObjectList();


    static DataFetcher idFetcher = environment -> {
        return "id"; // R2D2
    };


    static DataFetcher stringFetcher = environment -> {
        return "Hello World!"; // R2D2
    };


    static DataFetcher listOfStringFetcher = environment -> {
        return SchemaWiring.listOfString; // R2D2
    };


    static DataFetcher listOfObjectFetcher = environment -> {
        return SchemaWiring.listOfObject; // R2D2
    };

    static TypeResolver typeResolver = environment -> {
        return (GraphQLObjectType) environment.getSchema().getType("Query");
    };

    private static List<String> generateStringList() {
        List<String> s = new ArrayList<>();
        for (int i = 0; i < 100; i++) {
            s.add("Hello World!");
        }
        return s;
    }
    private static List<Object> generateObjectList() {
        List<Object> s = new ArrayList<>();
        for (int i = 0; i < 100; i++) {
            s.add("");
        }
        return s;
    }
}
