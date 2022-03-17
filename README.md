# poc-cuelang
proof of concept to try using cuelang for plugin management

This POC is the opportunity for us to answser a certain amount of questions we have today in Perses:

* Will we be able to inject new panel schemas at runtime ?
* Once a new schema is injected, is it considered by the backend and it is able to accept new data ?
* Should we use Cuelang for all our objects or just for the dashboards where we want to inject new panels ?
* How Cuelang and Jsonnet interact ? 
* We plan to have Perses native on k8s working with CRDs. How will it work with Cuelang ? It is a concurrent technology ?
