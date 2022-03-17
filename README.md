# poc-cuelang

This is a proof of concept to try using [CUE](https://cuelang.org/) for plugin management.

This POC is the opportunity for us to answer a certain amount of questions we have today in Perses :

* Will we be able to inject new panel schemas at runtime ?
* Once a new schema is injected, is it considered by the backend and it is able to accept new data ?
* Should we use CUE for all our objects or just for the dashboards where we want to inject new panels ?
* How CUE and Jsonnet interact ?
* We plan to have Perses native on k8s working with CRDs. How will it work with CUE ? It is a concurrent technology ?
