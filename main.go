package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/golang/protobuf/proto"
	plugin_go "github.com/golang/protobuf/protoc-gen-go/plugin"
)

type message map[string]interface{}

func (m *message) Range() map[string]interface{} {
	return *m
}

func merge(prefix string, m message) map[string]interface{} {
	temp := make(map[string]interface{})

	for k, v := range m.Range() {
		temp[prefix+"."+k] = v
	}

	return temp
}

func run(request *plugin_go.CodeGeneratorRequest) (*plugin_go.CodeGeneratorResponse, error) {
	resp := new(plugin_go.CodeGeneratorResponse)

	var (
		mess = make(map[string]*message)
	)

	for _, file := range request.ProtoFile {
		messageInFIle := make(map[string]*message)

		for _, m := range file.MessageType {
			rr := make(message)

			for _, f := range m.Field {
				if f.TypeName != nil {
					typeName := (*f.TypeName)[1:]

					for k, v := range merge(typeName, *mess[typeName]) {
						rr[k] = v
					}
				} else {
					rr[*f.Name] = f.Type.String()
				}
			}

			mess[*m.Name] = &rr
			messageInFIle[*m.Name] = &rr
		}

		b, _ := json.Marshal(messageInFIle)

		resp.File = append(resp.File, &plugin_go.CodeGeneratorResponse_File{
			Name:    proto.String(fmt.Sprintf("%v.py", *file.Name)),
			Content: proto.String(string(b)),
		})
	}

	return resp, nil
}

func main() {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err, "reading input")
	}

	req := new(plugin_go.CodeGeneratorRequest)
	if err := proto.Unmarshal(data, req); err != nil {
		log.Fatal(err, "parsing input proto")
	}

	res, err := run(req)

	if err != nil {
		log.Fatal(err)
	}

	// Send back the results.
	data, err = proto.Marshal(res)
	if err != nil {
		log.Fatal(err, "failed to marshal output proto")
	}
	_, err = os.Stdout.Write(data)
	if err != nil {
		log.Fatal(err, "failed to write output proto")
	}
}
