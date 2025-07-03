package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

func main() {
	var (
		kubeconfig   = flag.String("kubeconfig", filepath.Join(home(), ".kube", "config"), "path to kubeconfig file")
		namespace    = flag.String("namespace", "default", "the Kubernetes namespace")
		srcPod       = flag.String("src-pod", "", "the name of the source Pod")
		srcContainer = flag.String("src-container", "", "the name of the source container (optional)")
		srcPath      = flag.String("src-path", "", "the path in the source Pod to copy (file or directory)")
		dstPod       = flag.String("dst-pod", "", "the name of the destination Pod")
		dstContainer = flag.String("dst-container", "", "the name of the destination container (optional)")
		dstDir       = flag.String("dst-dir", "/", "the directory in the destination Pod to extract into")
	)
	flag.Parse()

	if *srcPod == "" || *srcPath == "" || *dstPod == "" {
		log.Fatalf("src-pod, src-path, and dst-pod are required")
	}

	// Build kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalf("failed to load kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to create clientset: %v", err)
	}

	if err := copyBetweenPods(
		context.Background(),
		clientset,
		config,
		*namespace,
		*srcPod, *srcContainer, *srcPath,
		*dstPod, *dstContainer, *dstDir,
	); err != nil {
		log.Fatalf("copy failed: %v", err)
	}

	fmt.Println("✅ Copy succeeded")
}

// copyBetweenPods streams a tar archive of srcPath in srcPod/srcContainer
// directly into dstPod/dstContainer under dstDir. Requires 'tar' in both containers.
func copyBetweenPods(
	ctx context.Context,
	clientset *kubernetes.Clientset,
	restConfig *rest.Config,
	namespace, srcPod, srcContainer, srcPath, dstPod, dstContainer, dstDir string,
) error {
	// common URL setup helper
	makeReq := func(pod, container string, cmd []string, stdin, stdout bool) (*rest.Request, error) {
		req := clientset.CoreV1().RESTClient().
			Post().
			Resource("pods").
			Namespace(namespace).
			Name(pod).
			SubResource("exec")

		opts := &corev1.PodExecOptions{
			Command:   cmd,
			Container: container,
			Stdin:     stdin,
			Stdout:    stdout,
			Stderr:    true,
		}
		req.VersionedParams(opts, parameterCodec())
		return req, nil
	}

	// 1) tar cf - srcPath  → stdout
	srcReq, err := makeReq(srcPod, srcContainer, []string{"tar", "cf", "-", srcPath}, false, true)
	if err != nil {
		return err
	}
	srcExec, err := remotecommand.NewSPDYExecutor(restConfig, "POST", srcReq.URL())
	if err != nil {
		return err
	}

	// 2) tar xf - -C dstDir ← stdin
	dstReq, err := makeReq(dstPod, dstContainer, []string{"tar", "xf", "-", "-C", dstDir}, true, true)
	if err != nil {
		return err
	}
	dstExec, err := remotecommand.NewSPDYExecutor(restConfig, "POST", dstReq.URL())
	if err != nil {
		return err
	}

	// Pipe the two together
	reader, writer := io.Pipe()
	defer reader.Close()

	// Run source exec in goroutine, writing its stdout into writer
	go func() {
		defer writer.Close()
		if err := srcExec.Stream(remotecommand.StreamOptions{
			Stdout: writer,
			Stderr: os.Stderr,
		}); err != nil {
			log.Printf("source exec error: %v", err)
		}
	}()

	// Run destination exec, reading from reader as its stdin
	return dstExec.Stream(remotecommand.StreamOptions{
		Stdin:  reader,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
}

// parameterCodec returns a codec for versioned params
func parameterCodec() runtime.ParameterCodec {
	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)
	return serializer.NewCodecFactory(scheme).ParameterCodec
}

// home returns the user home directory
func home() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
