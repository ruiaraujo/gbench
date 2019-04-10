(ns gbench.core
    (:require [ring.middleware.json :refer [wrap-json-response wrap-json-body]]
      [compojure.core :refer [POST defroutes]]
      [com.walmartlabs.lacinia :as lacinia]
      [gbench.schema :as schema]))


(defn handler [request]
      (response
        (lacinia/execute
          schema/test-schema (get-in request [:body "query"])
           nil nil)))

(defroutes my-routes
           (POST "/graphql" request (handler request)))

(def app
  (-> my-routes
      wrap-json-response
      wrap-json-body))
