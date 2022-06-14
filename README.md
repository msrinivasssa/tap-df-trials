# pythontrials

***Pre-req***

Install TAP in a K8s cluster using the `tap-pkg-values.yaml`

Create the cluster role binding, supply chain templates and supply chain under `./hack` directory
```
kapp deploy -a custom-supply-chain -f ./hack
```

Create the workload from `./config` directory
```
tanzu apps workload apply -f ./config/workload.yaml
```

***To upgrade version or upgrade install config***

Change the version in `tap-upgrade-values.yaml` to upgrade the TAP installation

Change the Install Parameters in `tap-pkg-values.yaml` to update install config
