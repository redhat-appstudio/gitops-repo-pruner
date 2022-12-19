#!/bin/sh

# Install Kubectl
cd /tmp
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"

# Get the list of all Application resources
chmod +x /tmp/kubectl
kubectl get applications --all-namespaces -o yaml > all-apps.yaml

echo "**** ALL-APPS"
cat all-apps.yaml

# Get the list of all GitOps repos
/pruner/gitops-repo-gc --operation list-all > all-repos.txt

echo "**** ALL-REPOS"
cat all-repos.txt
# Determine which gitops repositories need to be cleaned up
touch orphaned-repos.txt
while read p; do
    cat all-apps.yaml | grep $p > /dev/null
    if [ $? -ne 0 ]; then
        echo $p >> orphaned-repos.txt
    fi
done <all-repos.txt

echo "**** ORPHANED REPOS"
cat orphaned-repos.txt

# Delete the orphaned gitops repositories
while read p; do
    /pruner/gitops-repo-gc --operation delete-repo --repo $p
    if [ $? -ne 0 ]; then
        exit 1
    fi
done <orphaned-repos.txt