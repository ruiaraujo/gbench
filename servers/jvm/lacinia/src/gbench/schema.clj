(ns gbench.schema
    (:require [clojure.java.io :as io]
      [com.walmartlabs.lacinia.util :refer [attach-resolvers]]
      [com.walmartlabs.lacinia.parser.schema :as parser-schema]
      [com.walmartlabs.lacinia.schema :as schema]))


(def listOfStrings (map (fn [val] "Hello World!") (range 100)))
(def listOfObjects (map (fn [val] (schema/tag-with-type {} "Query")) (range 100)))

(def test-schema
  (schema/compile (parser-schema/parse-schema (slurp (io/resource "schema.gql"))
                                              {:resolvers {
                                                           :Query {
                                                                   :id               (fn [a b c] "id")
                                                                   :string           (fn [a b c] "Hello World!")
                                                                   :listOfStrings    (fn [a b c] listOfStrings)
                                                                   :listOfObjects    (fn [a b c] listOfObjects)
                                                                   :listOfInterfaces (fn [a b c] listOfObjects)}
                                                           }
                                               })))

