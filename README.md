# poc-cuelang

This is a proof of concept to try using [CUE](https://cuelang.org/) for plugin management.

This POC is the opportunity for us to answer a certain amount of questions we have today in Perses :

* Will we be able to inject new panel schemas at runtime?
  * -> **Yes**, thanks to a combinination of CUE & [fsnotify](https://github.com/fsnotify/fsnotify).
* Once a new schema is injected, is it considered by the backend and it is able to accept new data?
  * -> **Yes**, if the new schema meets the base requirements we expect from any panel schemas it is accepted and ready to be used.
* Should we use CUE for all our objects or just for the dashboards where we want to inject new panels?
  * -> **TBD**
* Could CUE interact with Jsonnet? How?
  * -> **Yes they could**. To manage config as code (= main purpose of CUE), Jsonnet and CUE are actually concurrent technologies. Jsonnet and CUE both originate from Google, and CUE was actually developped as an answer to the pitfalls of the languages like Jsonnet. More info [here](https://github.com/cue-lang/cue/discussions/669) and [here](https://github.com/cue-lang/cue/issues/33#issuecomment-873385381). However, in our case where we use CUE for [data validation](https://cuelang.org/docs/usecases/validation/), the dashboards generation (on user side) and validation (on Perses server side) are decorrelated processes, thus the generation part could be done with whatever technology suits you best (Jsonnet, CUE..). One advantage of using CUE for the dashboard generation is that people could reuse/import our schema definition files, to take advantage of the checks defined in those. However this would not really help to factorize/template things.
* We plan to have Perses native on k8s working with CRDs. How will it work with CUE? It is a concurrent technology?
  * -> **TBD**. It would work anyway because as said in the previous point, generation is decorrelated from validation. However it's not certain yet how we could avoid a duplication & thus a misalignment in the checks between the CRD & the CUE def. Should we (partially?) generate the CRD from the CUE definition ? Or, as for the dashboard go struct, fully delegate the panel validation to CUE thus accept any value for the panels at CRD level ? To be explored.
