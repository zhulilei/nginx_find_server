package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type nginxslice []string

var (
	records     nginxslice
	find        int
	server_name = flag.String("S", "", "server_name")
	//file_name   = flag.String("F", "", "filename_name")
	path_name     = flag.String("P", "", "path_name")
	upstream_name = flag.String("U", "", "upstream_name")
	with_proxy    = flag.Bool("A", false, "false")
	proxy_array   []string
)

func main() {
	flag.Parse()
	fileList := strings.Split(getFilelist(*path_name), "\r\n")

	if *server_name != "" && *upstream_name != "" {
		fmt.Println(*path_name, *upstream_name)
		panic("不要同时使用 域名查询和后端查询，只能每次使用一个")
	}

	if *server_name != "" {
		find = 0
	} else {
		find = 1
	}

	for i := 0; i < len(fileList)-1; i++ {
		if asFile(fileList[i], *server_name, *upstream_name) {
			fmt.Printf("\n %c[1;40;32m%s%c[0m\n\n", 0x1B, fileList[i], 0x1B)
			//fmt.Printf(fileList[i] + "\r\n")
		}
	}

}

func asFile(file_name, server_name, upstream_name string) bool {
	result := false
	records = []string{""}
	getStrings(file_name)
	matchIndexSlice := records.matchName(server_name, upstream_name)

	for _, matchIndex := range matchIndexSlice {
		startIndex := records.findStart(matchIndex)
		if startIndex == 0 {
			fmt.Println("hehe")
			continue
		}
		//fmt.Println("startIndexis", startIndex)

		endIndex := records.findEnd(startIndex)
		//fmt.Println("endIndexis", endIndex)

		if endIndex == 0 {
			continue

		}

		switch find {
		case 0:
			for index := startIndex - 1; index < endIndex; index++ {
				fmt.Printf(records[index])
				proxy := getProxyPass(records[index])
				if proxy != "" {
					proxy_array = append(proxy_array, proxy)
				}
				result = true
			}

			//fmt.Println(proxy_array)
			//fmt.Println("over")
			for _, value := range proxy_array {
				//fmt.Println("%d: %s", i, value)
				if *with_proxy != false {
					listProxyPass(file_name, value)
				}
			}

		case 1:
			for index := startIndex; index < endIndex; index++ {
				fmt.Printf(records[index])
				result = true
			}

		}

	}

	return result

}
func CheckErr(err error) {
	if nil != err {
		panic(err)
	}
}

func listProxyPass(file_name, upstream string) {
	find = 1
	asFile(file_name, "", upstream)
	asFile("/etc/nginx/conf.d/upstream.conf", "", upstream)

}

func getStrings(filename string) nginxslice {
	f, err := os.Open(filename)
	CheckErr(err)
	defer f.Close()

	reader := bufio.NewReader(f)
	for {
		line_context, err := reader.ReadString('\n') //以'\n'为结束符读入一行

		if err != nil || io.EOF == err {
			break
		}
		records = append(records, line_context)
	}

	return records

}

func (records nginxslice) matchName(server_name string, upstream_name string) (matchIndexSlice []int) {
	switch find {
	case 0:
		for i, v := range records {
			if strings.Contains(v, "server_name") {
				if strings.Contains(v, " "+server_name+";") || strings.Contains(v, " "+server_name+" ") {
					matchIndexSlice = append(matchIndexSlice, i+1)
				}
			}

		}
	case 1:
		for i, v := range records {
			if strings.Contains(v, "upstream") {
				if strings.Contains(v, " "+upstream_name+" ") || strings.Contains(v, " "+upstream_name+"{") {
					matchIndexSlice = append(matchIndexSlice, i)
				}
			}

		}

	}
	//fmt.Println(matchIndexSlice)

	return matchIndexSlice

}

func getProxyPass(line string) (proxy string) {
	if strings.Contains(line, "proxy_pass") {
		proxyArray := strings.Split(line, "//")
		proxy = strings.Replace(proxyArray[len(proxyArray)-1], ";", "", -1)
		proxy = strings.Replace(proxy, "\n", "", -1)
		//fmt.Println("proxyArray is", proxy)
		return proxy
	}
	return
}

func (records nginxslice) findStart(matchIndex int) (startIndex int) {
	switch find {
	case 0:
		for startIndex := matchIndex; startIndex >= 0; startIndex-- {
			if strings.Contains(records[startIndex], "server") && strings.Contains(records[startIndex], "{") {
				//fmt.Println(startIndex + 1)
				return startIndex + 1
			}
		}
	case 1:
		return matchIndex
	}

	return
}

func (records nginxslice) findEnd(startIndex int) (endIndex int) {
	brace := make(map[string]int)
	switch find {
	case 0:
		for endIndex := startIndex - 1; endIndex < len(records); endIndex++ {
			if strings.Contains(records[endIndex], "{") {
				brace["left"]++
			}
			if strings.Contains(records[endIndex], "}") {
				brace["right"]++
			}
			if brace["left"] != 0 && brace["right"] != 0 && brace["left"] == brace["right"] {
				return endIndex + 1
			}
		}
	case 1:
		for endIndex := startIndex; endIndex < len(records); endIndex++ {
			if strings.Contains(records[endIndex], "{") {
				brace["left"]++
			}
			if strings.Contains(records[endIndex], "}") {
				brace["right"]++
			}
			if brace["left"] != 0 && brace["right"] != 0 && brace["left"] == brace["right"] {
				return endIndex + 1
			}
		}
	}

	//fmt.Println("brace is", brace)
	return
}

//##################### start TraversalDir #################

func getFilelist(path string) string {
	var strRet string
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		strRet += path + "\r\n"
		return nil
	})
	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
	}
	return strRet
}
