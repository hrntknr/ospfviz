package main

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func configParser(fd *os.File) error {
	reader := bufio.NewReaderSize(fd, 4096)

	routers := []Router{}

	var line string

	var next = true
	var initRouter = true
	var initLink = false
	var returnCount = 0
	var marginReturn = 2
	var isRouter = true

	var linkType LinkType
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
			if returnCount == 1 && !isRouter {
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
				isRouter = false
				next = false
				initLink = false
				linkTypeStr := strings.Trim(strings.Split(line, ":")[1], " ")
				switch linkTypeStr {
				case "a Transit Network":
					linkType = TransitNetwork
					break
				case "Stub Network":
					linkType = StubNetwork
					break
				case "another Router (point-to-point)":
					linkType = P2PNetwork
					break
				default:
					return fmt.Errorf("invalid network type: %s", linkTypeStr)
				}
			}
			if initRouter {
				// start router
				router = Router{}
				isRouter = true
				next = false
				initRouter = false
			}
			if returnCount == 0 {
				// continuous line
				if isRouter {
					setAttr(reflect.ValueOf(&router), line, "vyos")
				} else {
					switch linkType {
					case TransitNetwork:
						link.Type = TransitNetwork
						setAttr(reflect.ValueOf(&link.Transit), line, "vyos")
						break
					case StubNetwork:
						link.Type = StubNetwork
						setAttr(reflect.ValueOf(&link.Stub), line, "vyos")
						break
					case P2PNetwork:
						link.Type = P2PNetwork
						setAttr(reflect.ValueOf(&link.P2P), line, "vyos")
						break
					}
				}
			}
			if returnCount == 1 {
				// end link space
				next = false
				initLink = true
			}
			if returnCount >= 2 {
				// end router space
				next = false
				initRouter = true
			}
			returnCount = 0
		}
	}
	fmt.Printf("%+v\n", routers)
	return nil
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
