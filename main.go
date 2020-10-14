package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/drgrib/alfred"
)

type TitleDesc struct {
	title string
	desc  string
}

func winnowtitle(title []byte, args []string) []byte {
	titlefields := bytes.Split(title, []byte(", "))

	if len(titlefields) == 1 {
		return titlefields[0]
	}

	for _, t := range titlefields {
		for _, a := range args {
			if bytes.HasPrefix(t, []byte(a)) {
				return t
			}
		}
	}

	return []byte{}
}

func mkpath() string {
	// TODO(rjk): Handle the error case nicely.
	h, _ := os.UserHomeDir()

	p := []string{
		"/bin",
		"/usr/bin",
		"/usr/local/bin",
		filepath.Join(h, ".ghcup/bin"),
		filepath.Join(h, ".cargo/bin"),
		filepath.Join(h, "bin"),
	}

	p9 := os.ExpandEnv("$PLAN9")
	if p9 != "" {
		p = append(p, filepath.Join(p9, "bin"))
	}

	return strings.Join(p, ":")
}

func main() {
	log.Println("hi")
	log.Println("args", os.Args)

	cmd := exec.Command("/usr/bin/whatis", os.Args[1:]...)

	// Need so that Plan9 implementation doesn't mess it up, I have
	// to reconstruct the PATH env.
	// TODO(rjk): When there are commands that are different between MacOS and
	// Plan9 but with identical names, we will end up with the wrong page.
	cmd.Env = append(cmd.Env, "PATH="+mkpath())

	log.Println(cmd.Env)

	out, err := cmd.Output()
	if err != nil {
		// Generate something else probably?
		log.Println("cmd.Output went wrong:", err)
		os.Exit(1)
	}
	// log.Println("outputs", string(out))

	lines := bytes.Split(out, []byte("\n"))
	tds := make([]*TitleDesc, 0, len(lines))

	for _, v := range lines {
		log.Println("oneline ", string(v))
		if bytes.HasSuffix(v, []byte(": nothing appropriate")) {
			tds = append(tds, &TitleDesc{
				title: string(bytes.TrimSuffix(v, []byte(": nothing appropriate"))),
				desc:  "missing whatis",
			})
			log.Println("nothing appropriate ", string(v))
			continue
		}

		cell := bytes.SplitN(v, []byte(" - "), 2)
		if len(cell) > 0 {
			cell[0] = winnowtitle(cell[0], os.Args[1:])
		}
		log.Println("split the line", len(cell))

		switch {
		case len(cell) < 1:
			log.Println("0 cells")
			continue
		case len(cell) < 2 && len(cell[0]) > 0:
			log.Println("1 cells", string(cell[0]))
			tds = append(tds, &TitleDesc{
				title: string(bytes.TrimSpace(cell[0])),
			})
		case len(cell) < 3 && len(cell[0]) > 0 && len(cell[1]) > 0:
			log.Println("2 cells", string(cell[0]), string(cell[1]))
			tds = append(tds, &TitleDesc{
				title: string(bytes.TrimSpace(cell[0])),
				desc:  string(bytes.TrimSpace(cell[1])),
			})
		}
	}

	for _, v := range tds {
		log.Println(v.title, " --> ", v.desc)
	}

	for _, v := range tds {
		alfred.Add(alfred.Item{
			Title:        v.title,
			Subtitle:     v.desc,
			Arg:          v.title,
			UID:          v.title,
			Autocomplete: v.title,
		})
	}

	if len(tds) == 0 && len(os.Args) > 0 {
		t := string(os.Args[1])
		alfred.Add(alfred.Item{
			Title:        t,
			Subtitle:     t,
			Arg:          t,
			UID:          t,
			Autocomplete: t,
		})
	}

	alfred.Run()
}
