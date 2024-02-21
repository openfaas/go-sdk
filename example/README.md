### Examples
Folder containes examples about how to use [openfaas](https://www.openfaas.com/) `go-sdk` API for different supported operations. Each subfolder contains a `main.go` file which can be executed independently. Before running them, please expose relevant environment variables such as `OPENFAAS_USERNAME`, `OPENFAAS_PASSWORD` and `OPENFAAS_GATEWAY_URL`. Do note, these are not standard environment variables. You can change them in your usecase.

Below is list of examples
1. [Deploy Function](./deploy-function/main.go)
2. [Update Function](./update-function/main.go)
3. [Scale Function](./scale-function/main.go)
4. [Get All Functions Of A Namespace](./get-functions/main.go)
5. [Create Namespace](./create-namespace/main.go)
6. [Update Namespace](./update-namespace/main.go)
7. [Get All Namespaces](./get-namepsaces/main.go)
8. [Create Secret](./create-secret/main.go)
9. [Update Secret](./update-secret/main.go)
10. [Get Logs Of A Function](./logs/main.go)



