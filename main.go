package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"

	"github.com/drgrib/alfred"	
)

type TitleDesc struct {
	title string
	desc string
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


func main() {
	log.Println("hi")
	log.Println("args", os.Args)

	cmd := exec.Command("/usr/bin/whatis", os.Args[1:]...)

	// Need so that Plan9 implementation doesn't mess it up.
	cmd.Env = append(cmd.Env, "PATH=/usr/bin:/bin")
	
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
		cell := bytes.Split(v, []byte(" - "))

		if len(cell) > 0 {
			cell[0] = winnowtitle(cell[0], os.Args[1:])
		}
		
		switch {
		case len(cell) < 1:
			continue
		case len(cell) < 2 && len(cell[0]) > 0:
			tds = append(tds, &TitleDesc{
				title: string(bytes.TrimSpace(cell[0])),
			})
		case len(cell) < 3 && len(cell[0]) > 0 && len(cell[1]) > 0: 
			tds = append(tds, &TitleDesc{
				title: string(bytes.TrimSpace(cell[0])),
				desc: string(bytes.TrimSpace(cell[1])),
			})
		}
	}

	for _, v := range tds {
		log.Println(v.title, " --> ", v.desc)
	}

	for _, v := range tds {
		alfred.Add(alfred.Item{
		Title:    v.title,
		Subtitle: v.desc,
		Arg:      v.title,
		UID:      v.title,
		Autocomplete: v.title,
		})
	}
	
	alfred.Run()
}
