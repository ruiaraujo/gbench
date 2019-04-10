(defproject gbench "0.1.0-SNAPSHOT"
            :dependencies [[org.clojure/clojure "1.9.0-beta2"]
                           [ring/ring-jetty-adapter "1.6.2"]
                           [ring/ring-json "0.4.0"]
                           [compojure "1.6.0"]
                           [com.walmartlabs/lacinia "0.32.0"]]
            :plugins [[lein-ring "0.12.1"]]
            :ring {:handler gbench.core/app})
