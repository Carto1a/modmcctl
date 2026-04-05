package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func main() {
	mode, _ := Flag("mode", "client", "client|server|both", Contains("not supported mode", "client", "server", "both"))
	clientDirFlag, _ := Flag("client-dir", "", ".minecraft client absolute path", nil)
	serverDirFlag, _ := Flag("server-dir", "", ".minecraft server absolute path", nil)
	version := Required(Flag("version", "", "minecraft version", nil))
	loader := Required(Flag("loader", "", "neoforge|fabric", Contains("not supported loader", "neoforge", "fabric")))
	providerName, _ := Flag("provider", "modrinth", "modrinth|curseforge", Contains("not supported provider", "modrinth", "curseforge"))
	modsFlag := Required(Flag("mods", "", "list of mods slugs/names, separated by comma", nil))
	flag.Parse()
	Validate()

	clientDir := getClientDir(*clientDirFlag)
	serverDir := getServerDir(*serverDirFlag)
	modsDirs := make([]string, 0)

	if *mode == "client" || *mode == "both" {
		modsDirs = append(modsDirs, filepath.Join(clientDir, "mods"))
	}
	if *mode == "server" || *mode == "both" {
		modsDirs = append(modsDirs, filepath.Join(serverDir, "mods"))
	}

	if err := installLoader(*loader, *mode, clientDir, serverDir, *version); err != nil {
		fmt.Println("error: ", err)
		os.Exit(1)
	}

	var provider ModProvider
	if *providerName == "curseforge" {
		provider = &CurseForgeProvider{}
	} else {
		provider = &ModrinthProvider{}
	}

	if *modsFlag != "" && provider != nil {
		modList := strings.Split(*modsFlag, ",")
		var wg sync.WaitGroup
		sem := make(chan struct{}, 5)

		for _, slug := range modList {
			slug = strings.TrimSpace(slug)
			if slug == "" { continue }

			wg.Add(1)
			go func(s string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				url, filename, err := provider.FetchMod(s, *version, *loader)
				if err != nil {
					fmt.Println("error: ", err)
					return
				}
				for _, modsDir := range modsDirs {
					ensureDir(modsDir)
					path := filepath.Join(modsDir, filename)
					if _, err := os.Stat(path); err == nil {
						fmt.Println("skip: ", filename)
						return
					}

					fmt.Println("downloading mod: ", filename)
					downloadFile(url, path)
				}

			}(slug)
		}
		wg.Wait()
	}
}
