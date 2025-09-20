Local Kubernetes (k3d) with ArgoCD, Local Registry & Chart Repo Setup
Overview and Tooling
For a smooth local development environment, you can combine k3d (to run a lightweight K3s cluster in Docker) with ArgoCD, a local container image registry, and a local Helm chart repository. This setup lets you iterate on manifests, container images, and Helm charts entirely on your Linux machine without external dependencies. We will use existing proven tools (k3d, ArgoCD, Docker registry, ChartMuseum, Helm) and minimal configuration to achieve the following:
• ArgoCD with Local Manifests – ArgoCD is deployed in the k3d cluster and watches a folder on your disk for Kubernetes manifests (no remote Git server required).
• Local Container Image Registry – A Docker registry running locally (in Docker) is integrated with the k3d cluster so you can push images from your machine and have the cluster pull them directly 1 .
• Local Helm Chart Repository – A ChartMuseum service (Helm chart repository) runs locally so you can push packaged Helm charts and have ArgoCD deploy them as needed 2 .
This approach uses standard tools and ensures everything works on Linux, keeping the development loop fast and self-contained. Below is a breakdown of the setup and brief guidance for each component.
Setting Up k3d Cluster with Local Registry
First, create a local Kubernetes cluster with k3d and include a private Docker registry for images. k3d can automatically start a registry container and configure the cluster’s container runtime to trust and use it:
• Create a registry and cluster: For example, you can run:
k3d registry create myregistry.localhost --port 5000
k3d cluster create devcluster --registry-use k3d-myregistry.localhost:
5000
The first command starts a local registry (listening on host port 5000), and the second creates a k3s cluster named devcluster that knows about this registry 3 . k3d will automatically inject the registry info into the cluster’s configuration (the registries.yaml for containerd), so cluster pods can pull images from myregistry.localhost:5000 without extra TLS or auth setup 1 .
• Push local images: Build your Docker images and tag them with the local registry’s address (for example, myregistry.localhost:5000/my-app:latest ). You can then push them from your host machine with docker push . Because k3d set up the cluster’s DNS/resolver, the
1
name myregistry.localhost will resolve inside the cluster, and kubelet/containerd will pull from the local registry when deploying pods 4 5 . (If using a \*.localhost domain for the registry name, ensure your Linux host can resolve it to 127.0.0.1 – e.g. via the
nss-myhostname mechanism as noted in k3d docs 6 .)
• Expose cluster services (optional): If you need to access cluster services (like ArgoCD UI or ChartMuseum) from your host, use k3d’s port mapping. For example, add -p "8080:80@loadbalancer" to expose the k3d ingress (or specify NodePorts) as needed. This will be useful to reach ArgoCD’s web UI or ChartMuseum from your browser.
At this stage, you have a single-node Kubernetes cluster (by default k3d uses one server node) with an integrated registry. Using a single node simplifies volume mounting for ArgoCD (described next) because all pods schedule on the one node by default.
Deploying ArgoCD with a Local Manifests Folder
Next, install ArgoCD into the k3d cluster (for example, apply the official ArgoCD manifests in a namespace like argocd ). To have ArgoCD use a local disk folder as the source of truth for manifests, we will mount that folder into the ArgoCD repo-server pod and register it as a file-based repo:
• Mount the local repo folder into the cluster: Suppose your Kubernetes manifests are in a local directory (e.g. ~/workspace/my-app-config/ ). You can mount this folder into the k3d node and ArgoCD. The simplest way is to use k3d’s volume mount at cluster creation. For example:
k3d cluster create devcluster -v ~/workspace/my-app-config:/data/my- app@server:0 . This maps your host folder into the K3s container (node) at /data/my-app . Then, in ArgoCD, define a PersistentVolume that uses a hostPath referring to
/data/my-app on the node, and a corresponding PersistentVolumeClaim in the argocd namespace. Patch the ArgoCD argocd-repo-server deployment to mount that PVC at a known path (say /repo inside the repo-server container). For instance, you might add to the deployment spec:
volumeMounts: - mountPath: /repo
name: local-repo
readOnly: true
volumes: - name: local-repo
persistentVolumeClaim:
claimName: local-repo-pvc
This ensures the ArgoCD repo-server container can see your host files under /repo (read-only) 7 . (The Stack Overflow example of this setup mounted a host path and PVC to /tmp/local_repo in argocd-repo-server 8 – the exact path is not important as long as it matches what you use in the next step.)
• Register the file-based repo in ArgoCD: ArgoCD does not natively browse arbitrary local folders, but it can treat a file:// URL as a repository source if the repo-server has access to it. Ensure your local manifest folder is a valid git repository (initialize it with git init and commit your YAML manifests, so it has a .git directory). Then you can create an ArgoCD
2

Application pointing to that repo using a file URL. For example:
argocd app create local-app \
 --repo file:///repo \
 --path . \
 --dest-server https://kubernetes.default.svc \
 --dest-namespace default --sync-policy auto
In the above, file:///repo corresponds to the mount path inside the repo-server. ArgoCD will treat this like a normal Git repository, just accessed via the local filesystem. (Behind the scenes, ArgoCD’s repo-server will perform a local git operation on that path.) This approach has been used to run ArgoCD with a local git repo in k3d clusters 9 . By using
--sync-policy auto (or setting auto-sync in the Application spec), ArgoCD will continuously sync the cluster to match changes you make in the local manifests folder. Whenever you update a manifest file in that folder and (re)commit, ArgoCD can detect the change and apply it, all without any remote Git service 10 11 .
Note: When using file:// repositories, be sure the ArgoCD repo-server runs with permissions to read the mounted files. In k3d, files may be owned by root in the node container. If needed, adjust permissions or run the repo-server container as root, or mount with appropriate UID/GID, so ArgoCD can read the repo content 12 . Also, restrict this setup to local/offline development – for production, a Git server is preferred (ArgoCD is designed for GitOps).
Running a Local Helm Chart Repository (ChartMuseum)
For Helm charts, you can deploy ChartMuseum to act as a local chart repository. ChartMuseum is a lightweight Helm chart server that stores charts on disk (or in cluster storage) and provides an HTTP API to upload and index charts. It’s an ideal choice here because it runs locally and gives you full control over chart versions and availability 2 . Set it up as follows:
• Deploy ChartMuseum: You can run ChartMuseum in the k3d cluster (e.g., in a dedicated namespace). There is an official Helm chart and Docker image for ChartMuseum 13 . For a quick start, you might run it as a Deployment with a Service. For example, use the chartmuseum/ chartmuseum image and configure it to use a PVC for storage. Expose it on a port accessible to your host (perhaps via NodePort or an Ingress). If using NodePort, you can map that port to localhost using k3d (similar to how we exposed ArgoCD).
• Push charts locally: Develop your Helm charts on your machine, then package them with helm package . You can push charts to ChartMuseum either by using the Helm push plugin or via HTTP API. For instance, if ChartMuseum is listening at http://localhost:8080 (or an
appropriate URL/port), you can do:
helm package ./mychart -d ./packages
curl --data-binary "@mychart-0.1.0.tgz" http://localhost:8080/api/charts
This will upload your packaged chart to the local repository (ChartMuseum will index it) 14 . Alternatively, the helm-push plugin allows helm push mychart-0.1.0.tgz repo-name directly if configured.
3

• Use charts in ArgoCD: ArgoCD supports Helm chart repositories (including ChartMuseum) as sourcesforapplications 15 .You’llneedtotellArgoCDaboutthisHelmrepo.Thesimplestwayis to add it via ArgoCD’s CLI or UI: for example, argocd repo add http:// chartmuseum.chartmuseum.svc.cluster.local:8080 --name local-charts --type helm . (Use the appropriate URL that ArgoCD can reach – if ArgoCD is running in the cluster, the internal service address is reachable; if not, use the NodePort/ingress address.) ArgoCD requires explicitly adding Helm repos in its configuration because they aren’t public by default 16 . Once added, you can create ArgoCD Applications that reference this repo and a specific chart name and version. For example, an Application spec might have:
spec:
source:
repoURL: "http://chartmuseum.chartmuseum.svc.cluster.local:8080"
chart: mychart
targetRevision: "0.1.0"
ArgoCD will pull the mychart-0.1.0.tgz from ChartMuseum and deploy it (using Helm under the hood). Whenever you push a new chart version, you can update the targetRevision (or use semantic version ranges if supported) and ArgoCD will sync to the new chart release.
Developer Workflow and Considerations
With the above components, your local workflow would be as follows:
• Develop Kubernetes manifests or Helm charts on your local machine. For plain manifests, commit changes in the local git folder that ArgoCD is watching. For Helm charts, package and push updates to ChartMuseum.
• Build Docker images for your apps and push them to the local registry
( myregistry.localhost:5000 ). The Kubernetes manifest files or Helm chart values should reference the images by this local registry address. For example, set your Pod/Deployment specs image to myregistry.localhost:5000/your-app:tag . Because the cluster trusts this registry, image pulls will succeed 17 18 .
• Let ArgoCD deploy and reconcile the changes. ArgoCD will either apply the new manifest YAML from your folder or perform a Helm upgrade to apply the new chart, ensuring the cluster reflects the state of your local source. You get the benefits of GitOps-style sync (automatic deployments, drift detection) without requiring an external Git host.
This solution uses readily available tools and minimal custom scripting. k3d handles the Kubernetes and registry setup, ArgoCD handles continuous delivery from a local source, and ChartMuseum manages Helm charts locally. All components run on Linux and are containerized, making the environment reproducible and easy to tear down or spin up. It’s a practical way to achieve a fully local GitOps-like setup:
• No external dependencies: No need for a remote Git repository or Docker registry – everything is on the local disk or localhost network. This is useful for offline scenarios or rapid local iteration
10 .
• Leverage proven software: ArgoCD for syncing manifests, Docker registry for images,
ChartMuseum for charts – these are industry-standard tools, so you avoid reinventing the wheel. Each is designed to work on Linux and in containerized setups.
4

• Easy cleanup and reset: Since k3d runs the cluster and registry in Docker containers, you can remove the cluster ( k3d cluster delete ) to clean up. Your local manifests folder and any chart packages remain on disk for source control or next iteration.
If a one-click “full solution” script isn’t available, the above combination is the most reliable path to meet all the requirements. In summary, use k3d with a built-in registry, install ArgoCD with a mounted local git folder, and run ChartMuseum for charts. This provides a robust local Kubernetes environment where developers can build, push, and deploy applications using only local resources – achieving the desired GitOps workflow without needing remote services.
Sources:
• k3d Documentation – Using a Local Registry 19 3
• Example of ArgoCD with a file-based repo in k3d 9 7
• ChartMuseum for local Helm chart storage 2 14
• ArgoCD Docs – Support for Helm chart repositories like ChartMuseum 15
1 3 4 5 6 17 18 19 Using Image Registries - k3d https://k3d.io/v5.2.2/usage/registries/
2 13 14 Implementation of Multiple source Argo CD + Chartmuseum for 10 one-type microservices | by Loovatech | Medium https://medium.com/@loovatech/implementation-of-multiple-source-argo-cd-chartmuseum-for-10-one-type- microservices-212db5a13701
7 8 9 12 kubernetes - Run argocd in a k3d cluster with local repo - Stack Overflow https://stackoverflow.com/questions/77951224/run-argocd-in-a-k3d-cluster-with-local-repo
10 11 Use local path instead of a repository as source · argoproj argo-cd · Discussion #9912 · GitHub https://github.com/argoproj/argo-cd/discussions/9912
15 16 Repos - Argo CD - Declarative GitOps CD for Kubernetes https://kostis-argo-cd.readthedocs.io/en/first-page/basics/repositories/
(Helm chart repo in-cluster)
5
