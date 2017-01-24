package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func genDataset(nfiles, fsize, depth, ndirs int, dir string) error {
	fmt.Println("Generating dataset...")
	args := []string{
		"-files", fmt.Sprint(nfiles),
		"-filesize", fmt.Sprint(fsize),
		"-depth", fmt.Sprint(depth),
		"-dirs", fmt.Sprint(ndirs),
		dir,
	}
	cmd := exec.Command("random-files", args...)

	err := cmd.Run()
	if err != nil {
		return err
	}

	cmd2 := exec.Command("du", "-s", "-h")
	cmd2.Stdout = os.Stdout
	cmd2.Run()

	return nil
}

func initAndConfigNode(ipfsbin, dir string, reposync bool) error {
	cmd := exec.Command(ipfsbin, "init")
	cmd.Env = []string{"IPFS_PATH=" + dir}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command(ipfsbin, "bootstrap", "rm", "--all")
	cmd.Env = []string{"IPFS_PATH=" + dir}
	err = cmd.Run()
	if err != nil {
		return err
	}

	if !reposync {
		fmt.Println("setting nosync")
		cmd = exec.Command(ipfsbin, "config", "--json", "Datastore.NoSync", "true")
		cmd.Env = []string{"IPFS_PATH=" + dir}
		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func startDaemon(ipfsbin, workdir, routing string) (*os.Process, error) {
	args := []string{"daemon"}
	if routing != "" {
		args = append(args, "--routing="+routing)
	}

	cmd := exec.Command(ipfsbin, args...)
	cmd.Env = []string{"IPFS_PATH=" + workdir}
	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	time.Sleep(time.Second)
	return cmd.Process, nil
}

func doAdd(ipfsbin, workdir, dataset string, raw bool, chunker string) error {
	fmt.Println("starting add...")
	before := time.Now()
	args := []string{"add", "-r", "-q"}
	if raw {
		args = append(args, "--raw-leaves")
	}
	if chunker != "" {
		args = append(args, "--chunker", chunker)
	}
	args = append(args, dataset)
	cmd := exec.Command(ipfsbin, args...)
	cmd.Env = []string{"IPFS_PATH=" + workdir}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	took := time.Since(before)
	fmt.Println("add complete: ", took)
	return nil
}

func fatal(i interface{}) {
	fmt.Println(i)
	os.Exit(1)
}

func main() {
	routing := flag.String("routing", "", "specify routing type to use")
	reposync := flag.Bool("repo-sync", true, "flatfs datastore sync")
	chunker := flag.String("chunker", "", "chunker to use while adding")
	rawleaves := flag.Bool("raw-leaves", false, "use raw leaves for add")
	nfiles := flag.Int("num-files", 10, "random-files number of files per dir")
	fsize := flag.Int("file-size", 100000, "random-files filesize parameter")
	depth := flag.Int("dir-depth", 3, "random-files directory depth parameter")
	ndirs := flag.Int("num-dirs", 5, "random-files dirs parameter")
	dset := flag.String("to-add", "", "optionally specify data to test adding")
	ipfsbin := flag.String("ipfsbin", "", "ipfs binary to use")
	flag.Parse()

	var workdir string
	if len(flag.Args()) > 0 {
		workdir = flag.Args()[0]
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			fatal(err)
		}

		tempdir, err := ioutil.TempDir(cwd, "ipfs-test")
		if err != nil {
			fatal(err)
		}

		workdir = tempdir
	}

	if *ipfsbin == "" {
		ipfspath, err := exec.LookPath("ipfs")
		if err != nil {
			fatal(err)
		}
		*ipfsbin = ipfspath
	}

	var toadd string
	if *dset == "" {
		datasetpath := filepath.Join(workdir, "data")

		err := genDataset(*nfiles, *fsize, *depth, *ndirs, datasetpath)
		if err != nil {
			fmt.Println("gen dataset failed")
			fatal(err)
		}
	} else {
		toadd = *dset
	}

	ipfsdir := filepath.Join(workdir, ".ipfs")

	err := initAndConfigNode(*ipfsbin, ipfsdir, *reposync)
	if err != nil {
		fmt.Println("init node failed")
		fatal(err)
	}

	proc, err := startDaemon(*ipfsbin, ipfsdir, *routing)
	if err != nil {
		fmt.Println("start daemon failed")
		fatal(err)
	}
	fmt.Printf("daemon pid: %d\n", proc.Pid)

	err = doAdd(*ipfsbin, ipfsdir, toadd, *rawleaves, *chunker)
	if err != nil {
		fatal(err)
	}

	proc.Kill()
}
