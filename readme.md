# How to Use

The **k8s-search** tool allows you to search for specific string patterns within deployed Kubernetes Secrets and ConfigMaps across all namespaces in your cluster. It examines both the resource names and their content (in the `stringData` and `Data` fields) using a case-insensitive search.

Provide the search string using the `--string` flag. For example:

```bash
secret-search --string="my-pattern" [--verbose]
```

The optional `--verbose` flag enables additional logging to help debug or understand the search process.

---

## Eg

Below is an example of the tool's output when a match is found:

```
=== Scanning Kubernetes Secrets and ConfigMaps ===
Found secret in namespace: default; Secret: my-secret - password : supersecretpassword
Found config in namespace: production; configKey: apiEndpoint : https://api.example.com
```

---

## Exit Codes

The **k8s-search** tool uses exit codes to indicate the result of the search:

- **Exit Code 0:**  
  No matching Secrets or ConfigMaps were found. This indicates that the search string did not appear in any resource.
  
- **Exit Code 1:**  
  One or more matching Secrets or ConfigMaps were found. This signals that at least one instance of the search string was detected.

---

## Install

Download precompiled binaries for various platforms from the [releases page](benjaco/k8s-search/releases).

Alternatively, clone the repository and build the tool using Go:

```bash
git clone https://github.com/benjaco/k8s-search.git
cd k8s-search
go build -o secret-search .
```

