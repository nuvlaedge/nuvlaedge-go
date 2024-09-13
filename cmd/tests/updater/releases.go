package main

import "nuvlaedge-go/updater/release"

func main() {
	v := "2.15.1"

	n, err := release.GetNuvlaRelease(v)
	if err != nil {
		panic(err)
	}

	files := []string{"docker-compose.yml"}

	dFiles, err := n.GetComposeFiles(files, "/tmp/nuvlaedge/releases")
	if err != nil {
		panic(err)
	}

	for _, f := range dFiles {
		println(f)
	}
}
