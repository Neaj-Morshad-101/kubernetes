
# Pod to Pod File Copy Utility

This project demonstrates how to copy files and directories from one Kubernetes Pod to another without using `kubectl cp`. It uses the Kubernetes API to execute `tar` commands on the source and destination pods and streams the tar archive directly between them.

This is useful for scenarios where you need to perform automated backups, data migration, or state transfers between pods directly within a cluster, orchestrated by an external Go program.

## Prerequisites

- A running Kubernetes cluster (e.g., Minikube, kind, Docker Desktop).
- `kubectl` configured to communicate with your cluster.
- Go (version 1.18+ recommended).

## How It Works

The Go program (`main.go`) performs the following steps:
1. Connects to the Kubernetes cluster using a standard kubeconfig file.
2. Creates an `exec` session on the **source pod** to run `tar cf - <source_path>`. This command creates a tar archive of the specified path and writes it to standard output.
3. Creates an `exec` session on the **destination pod** to run `tar xf - -C <destination_dir>`. This command extracts a tar archive from its standard input into the specified directory.
4. The Go program then pipes the standard output of the source pod's command directly to the standard input of the destination pod's command.

This method is efficient as the data is streamed directly between the pods via the Kubernetes API server, without ever landing on the local filesystem where the Go program is running.

## Setup and Usage

Follow these steps to run the demonstration.

### 1. Clone the Repository

```sh
git clone <your-repo-url>
cd copy-files-between-pods
```

### 2. Create the Source and Destination Pods

The `manifests` directory contains the YAML for two simple `busybox` pods which will serve as our source (`pod1`) and destination (`pod2`).

Apply the manifests to your cluster:
```sh
kubectl apply -f manifests/
```

Verify that the pods are running:
```sh
$ kubectl get pods
NAME   READY   STATUS    RESTARTS   AGE
pod1   1/1     Running   0          30s
pod2   1/1     Running   0          30s
```

### 3. Generate and Copy Sample Data to the Source Pod

The `scripts/gen_sample_data.sh` script creates a directory with some sample files and subdirectories.

First, generate the sample data locally:
```sh
chmod +x gen_sample_data.sh
gen_sample_data.sh
```
This will create a `src_data` directory in your project root.

Now, copy this sample data into the source pod (`pod1`) at the path `/data`. We use `kubectl cp` for this initial setup step.
```sh
kubectl cp ./src_data pod1:/data/
```
> Note: The path inside the pod will be `/data/src_data`.

Verify the data exists in `pod1`:
```sh
$ kubectl exec pod1 -- ls -R /data
/data:
src_data

/data/src_data:
dir1
dir2
dir3
root_file1.txt
root_file2.csv
...and so on
```

### 4. Run the Go Program to Copy Data

Now we'll run the main Go program to copy the `/data/src_data` directory from `pod1` to the `/data` directory in `pod2`.

First, tidy the Go modules:
```sh
go mod tidy
```

Then, run the program. You can either build it or run it directly.

**Using `go run`:**
```sh
go run main.go \
  --src-pod=pod1 \
  --src-path=/data/src_data \
  --dst-pod=pod2 \
  --dst-dir=/
```


### 5. Verify the Copy

If the program ran successfully, you will see the output: `✅ Copy succeeded`.

Now, verify that the files exist in the destination pod (`pod2`):
```sh
$ kubectl exec pod2 -- ls -R /data
/data:
src_data

/data/src_data:
dir1
dir2
dir3
root_file1.txt
root_file2.csv

/data/src_data/dir1:
file1.log
sub1
sub2

/data/src_data/dir1/sub1:
deep_file1.json

/data/src_data/dir1/sub2:

/data/src_data/dir2:
sub1
sub2

/data/src_data/dir2/sub1:

/data/src_data/dir2/sub2:
text_data.txt

/data/src_data/dir3:
placeholder.txt
```
You should see the exact same file structure that was in `pod1`.

### 6. Clean Up

To delete the pods and other resources created during this tutorial, run:
```sh
kubectl delete -f manifests/
```