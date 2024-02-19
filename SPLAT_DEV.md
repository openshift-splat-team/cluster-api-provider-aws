# Building the Installer

1. Checkout https://github.com/openshift/installer/pull/7999/files
2. Build the installer to include CAPI assets:
   ```sh
   SKIP_TERRAFORM=y OPENSHIFT_INSTALL_CLUSTER_API=1 ./hack/build.sh   
   ```
3. Construct an `install-config.yaml` which enables the required feature gate
   ```yaml
   featureSet: CustomNoUpgrade
   featureGates:
   - ClusterAPIInstall=true
   ```
4. Install the cluster.
   Note: at this time no machines are created, only the CAPI and CAPA infrastructure manifests.  The AWSCluster resource is responsible for the creation of the public API load balancer.


# Experimenting with CAPA

The installer builds and embeds CAPI provider binaries.  Since these binaries are only rebuilt if the build process sees a change in CAPI/CAPA source code, the CAPA binary can be built separately and copied in to place.  This [branch](https://github.com/openshift-splat-team/cluster-api-provider-aws/tree/capa-eip-poc) is a poc which works with some of the controllers identified during research.

To build changes to the provider the CAPA binary is built and co pied in to the installer directory.  The installer is then built to incorporate the CAPA binary.

```sh
cp ../cluster-api-provider-aws/bin/manager cluster-api/bin/linux_amd64/cluster-api-provider-aws
SKIP_TERRAFORM=y OPENSHIFT_INSTALL_CLUSTER_API=1 ./hack/build.sh
```

```json
{

// See https://go.microsoft.com/fwlink/?LinkId=733558

// for the documentation about the tasks.json format

"version": "2.0.0",

"tasks": [

{

"label": "build",

"type": "shell",

"command": "cp ../cluster-api-provider-aws/bin/manager cluster-api/bin/linux_amd64/cluster-api-provider-aws; SKIP_TERRAFORM=y OPENSHIFT_INSTALL_CLUSTER_API=1 ./hack/build.sh",

"problemMatcher": [],

"dependsOn":[

"kill cluster API processes"

],

"group": {

"kind": "build",

"isDefault": true

}

},

{

"label": "kill cluster API processes",

"type": "shell",

"command": "killall etcd kube-apiserver cluster-api-provider-aws cluster-api || echo 0"

},

{

"label": "provision install-config",

"type": "shell",

"dependsOn": "build",

"command": "rm -r inst;mkdir inst;cp install-config.yaml inst/install-config.yaml"

},

{

"label": "create manifests",

"type": "shell",

"dependsOn": "provision install-config",

"command": "export OPENSHIFT_INSTALL_RELEASE_IMAGE_OVERRIDE=quay.io/openshift-release-dev/ocp-release:4.16.0-ec.2-x86_64; ./bin/openshift-install create manifests --dir inst"

},

{

"label": "install openshift",

"type": "shell",

"command": "export OPENSHIFT_INSTALL_RELEASE_IMAGE_OVERRIDE=quay.io/openshift-release-dev/ocp-release:4.16.0-ec.2-x86_64; ./bin/openshift-install create cluster --dir inst",

"dependsOn": "provision install-config"

}

]

}
```

