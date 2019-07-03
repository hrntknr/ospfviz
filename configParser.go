package main

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func configParser(fd *os.File) (*[]Router, error) {
	reader := bufio.NewReaderSize(fd, 4096)

	routers := []Router{}

	var line string

	var next = true
	var initRouter = true
	var initLink = false
	var returnCount = 0
	var marginReturn = 2
	var isLink = false

	var router Router
	var link Link

	for {
		if next {
			var err error
			line, err = reader.ReadString('\n')
			if err != nil {
				if marginReturn == 0 {
					break
				}
				line = "\n"
				marginReturn--
			}
			line = strings.Trim(line, "\n")
			line = strings.Trim(line, " ")
		} else {
			next = true
		}

		if line == "" {
			returnCount++
			if returnCount == 1 && isLink {
				// end link
				router.Links = append(router.Links, link)
			}
			if returnCount == 2 {
				// end router
				routers = append(routers, router)
			}
		} else {
			if initLink {
				// start link
				link = Link{}
				isLink = true
				next = false
				initLink = false
				linkTypeStr := strings.Trim(strings.Split(line, ":")[1], " ")
				switch linkTypeStr {
				case "a Transit Network":
					link.Type = TransitNetwork
					link.Transit = &TransitInfo{}
					break
				case "Stub Network":
					link.Type = StubNetwork
					link.Stub = &StubInfo{}
					break
				case "another Router (point-to-point)":
					link.Type = P2PNetwork
					link.P2P = &P2PInfo{}
					break
				default:
					return nil, fmt.Errorf("invalid network type: %s", linkTypeStr)
				}
			}
			if initRouter {
				// start router
				router = Router{}
				next = false
				initRouter = false
			}
			if returnCount == 0 {
				// continuous line
				if !isLink {
					setAttr(reflect.ValueOf(&router), line, "vyos")
				} else {
					switch link.Type {
					case TransitNetwork:
						setAttr(reflect.ValueOf(link.Transit), line, "vyos")
						break
					case StubNetwork:
						setAttr(reflect.ValueOf(link.Stub), line, "vyos")
						break
					case P2PNetwork:
						setAttr(reflect.ValueOf(link.P2P), line, "vyos")
						break
					}
				}
			}
			if returnCount == 1 {
				// end link space
				next = false
				initLink = true
				isLink = false
			}
			if returnCount >= 2 {
				// end router space
				next = false
				initRouter = true
			}
			returnCount = 0
		}
	}
	return &routers, nil
}

func setAttr(ref reflect.Value, line string, keyword string) error {
	words := strings.Split(line, ":")
	key := strings.Trim(words[0], " ")
	value := strings.Trim(words[1], " ")

	for i := 0; i < ref.Elem().Type().NumField(); i++ {
		field := ref.Elem().Type().Field(i)
		tag, ok := field.Tag.Lookup(keyword)
		if ok && tag == key {
			switch field.Type.Kind() {
			case reflect.String:
				ref.Elem().Field(i).Set(reflect.ValueOf(value))
				break
			case reflect.Int:
				num, err := strconv.Atoi(value)
				if err != nil {
					return err
				}
				ref.Elem().Field(i).Set(reflect.ValueOf(num))
				break
			}
		}
	}
	return nil
}
