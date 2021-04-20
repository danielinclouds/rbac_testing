

-----
# Scenario - multi-tenant cluster example
A cluster is shared by 2 teams: red and blue. 
- Both teams have multiple namespaces.
- User that don't belong to a red team should not be able to do any actions on red namespaces.

Blue team members:
 - dave@gmail.com
 - joe@gmail.com
Red team members:
 - mike@gmail.com

Blue team namespaces:
 - app1
 - app2
 - app3
Red team namespaces:
 - app4
 - app5


# Prerequisites
Create cluster
kind create cluster

Install rbac-manager
```sh
helm repo add fairwinds-stable https://charts.fairwinds.com/stable
helm repo update
kubectl create ns rbac-manager
helm install rbac-manager fairwinds-stable/rbac-manager -n rbac-manager
```

Install redis operator - this will show how to extend "edit" role
```sh
helm repo add ot-helm https://ot-container-kit.github.io/helm-charts/
helm repo update
kubectl create ns redis-operator
helm install redis-operator ot-helm/redis-operator -n redis-operator
```


# Extend edit cluster role
```sh
cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    rbac.authorization.k8s.io/aggregate-to-edit: "true"
  name: edit-redis
rules:
- apiGroups:
  - redis.redis.opstreelabs.in
  resources:
  - redis
  verbs:
  - '*'
EOF
```


# Define RBAC
Give edit permissions to all namespaces that belong to blue team
```sh
cat <<EOF | kubectl apply -f -
apiVersion: rbacmanager.reactiveops.io/v1beta1
kind: RBACDefinition
metadata:
  name: blue
rbacBindings:
  - name: blue
    subjects:
      - kind: User
        name: dave@gmail.com
      - kind: User
        name: joe@gmail.com
    roleBindings:
      - clusterRole: edit
        namespaceSelector:
          matchLabels:
            team: blue
EOF
```

Give edit permissions to all namespaces that belong to red team
```sh
cat <<EOF | kubectl apply -f -
apiVersion: rbacmanager.reactiveops.io/v1beta1
kind: RBACDefinition
metadata:
  name: red
rbacBindings:
  - name: red
    subjects:
      - kind: User
        name: mike@gmail.com
    roleBindings:
      - clusterRole: edit
        namespaceSelector:
          matchLabels:
            team: red
EOF
```


# Create namespaces
```
kubectl create ns app1
kubectl create ns app2
kubectl create ns app3

kubectl label ns app1 team=blue
kubectl label ns app2 team=blue
kubectl label ns app3 team=blue

kubectl create ns app4
kubectl create ns app5

kubectl label ns app4 team=red
kubectl label ns app5 team=red
```

# Run tests
Developer can create deployment 
```sh
kubectl auth can-i create deployment -n app1 --as=dave@gmail.com
# yes
```

Developer can't create role binding
```sh
kubectl auth can-i create rolebinding -n app1 --as=dave@gmail.com
# no
```

Developers can't change namespace labels
```sh
kubectl auth can-i label namespace -n app1 --as=dave@gmail.com
# no
```

dave@gmail.com from team blue can't list pods in team red namespace
```sh
kubectl auth can-i get pods -n app4 --as=dave@gmail.com
# no
```

Developer can't list pods in kube-system namespace
```sh
kubectl auth can-i get pods -n kube-system --as=dave@gmail.com
# no
```

Developer can create CRDs aggregated with "edit" cluster role
```sh
kubectl auth can-i create redis -n app1 --as=dave@gmail.com
# yes
```


# Run go tests
```sh
go test -v
```

