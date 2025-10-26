package main

import (
	"fmt"
	"strings"

	"github.com/jhump/protoreflect/desc/protoparse"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type ProtoParserOption func(*ProtoParser)

type ProtoParser struct {
	parser *protoparse.Parser

	Files map[string]protoreflect.FileDescriptor
}

// load proto content only from memory
func WithImportContents(contents map[string]string) ProtoParserOption {
	return func(p *ProtoParser) {
		p.parser.Accessor = protoparse.FileContentsFromMap(contents)
	}
}

func WithImportPaths(paths []string) ProtoParserOption {
	return func(p *ProtoParser) {
		p.parser.ImportPaths = paths
	}
}

func NewProtoParser(opts ...ProtoParserOption) *ProtoParser {
	p := &ProtoParser{
		parser: &protoparse.Parser{
			ImportPaths: []string{"."},
		},
		Files: make(map[string]protoreflect.FileDescriptor),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func NewDefaultProtoParser() *ProtoParser {
	return NewProtoParser()
}

func (p *ProtoParser) ParseProtoFile(filePath string) error {
	fileDesc, err := p.parser.ParseFiles(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse proto file %s: %v", filePath, err)
	}
	for _, fd := range fileDesc {
		p.Files[fd.GetName()] = fd.UnwrapFile()
	}
	return nil
}

func (p *ProtoParser) PrintProtoInfo() {
	for fileName, fileDesc := range p.Files {
		fmt.Printf("\n======== Proto File: %s ========\n", fileName)
		fmt.Printf("Syntax: %s\n", fileDesc.Syntax())
		fmt.Printf("Package: %s\n", fileDesc.Package())
		fmt.Printf("Path: %s\n", fileDesc.Path())

		imports := fileDesc.Imports()

		if imports.Len() > 0 {
			fmt.Println("Imports:")
			for i := 0; i < imports.Len(); i++ {
				fmt.Printf("  - %s\n", imports.Get(i).FullName())
			}
		}

		messages := fileDesc.Messages()
		if messages.Len() > 0 {
			fmt.Println("Messages:")
			for i := 0; i < messages.Len(); i++ {
				p.printMessage(messages.Get(i), 1)
			}
		}
		enums := fileDesc.Enums()
		if enums.Len() > 0 {
			fmt.Println("Enums:")
			for i := 0; i < enums.Len(); i++ {
				p.printEnum(enums.Get(i), 1)
			}
		}
		services := fileDesc.Services()
		if services.Len() > 0 {
			fmt.Println("Services:")
			for i := 0; i < services.Len(); i++ {
				p.printService(services.Get(i), 1)
			}
		}
	}
}

func (p *ProtoParser) printMessage(msg protoreflect.MessageDescriptor, indent int) {
	indentStr := strings.Repeat("    ", indent)
	fmt.Printf("%smessage %s {\n", indentStr, msg.Name())

	fields := msg.Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		fieldType := p.getFieldType(field)

		attr := '*'
		if field.HasOptionalKeyword() {
			attr = 'o'
		}
		if field.IsPacked() {
			attr = '+'
		}
		fmt.Printf("%s [%c]%s %s", indentStr, attr, fieldType, field.Name())

		if field.Kind() == protoreflect.MessageKind {
			fmt.Printf(";\n")
		} else if field.Kind() == protoreflect.EnumKind {
			fmt.Printf(" = %v;\n", field.DefaultEnumValue().Name())
		} else if field.Kind() == protoreflect.StringKind {
			fmt.Printf(" = \"%v\";\n", field.Default())
		} else {
			fmt.Printf(" = %v;\n", field.Default())
		}
	}

	msgs := msg.Messages()
	for i := 0; i < msgs.Len(); i++ {
		p.printMessage(msgs.Get(i), indent+1)
	}

	enums := msg.Enums()
	for i := 0; i < enums.Len(); i++ {
		p.printEnum(enums.Get(i), indent+1)
	}

	fmt.Printf("%s} \n", indentStr)
}

func (p *ProtoParser) printEnum(enum protoreflect.EnumDescriptor, indent int) {
	indentStr := strings.Repeat("    ", indent)
	fmt.Printf("%senum %s {\n", indentStr, enum.Name())

	enumValues := enum.Values()
	for i := 0; i < enumValues.Len(); i++ {
		value := enumValues.Get(i)
		fmt.Printf("%s  %s = %d;\n",
			indentStr,
			value.Name(),
			value.Number())
	}

	fmt.Printf("%s}\n", indentStr)
}

func (p *ProtoParser) printService(service protoreflect.ServiceDescriptor, indent int) {
	indentStr := strings.Repeat("    ", indent)
	fmt.Printf("%sservice %s {\n", indentStr, service.Name())

	methods := service.Methods()
	for i := 0; i < methods.Len(); i++ {
		method := methods.Get(i)
		fmt.Printf("%s  rpc %s (%s) returns (%s);\n",
			indentStr,
			method.Name(),
			method.Input().Name(),
			method.Output().Name())
	}
	fmt.Printf("%s}\n", indentStr)
}

func (p *ProtoParser) getFieldType(field protoreflect.FieldDescriptor) string {
	var typeStr string
	switch field.Kind() {
	case protoreflect.DoubleKind:
		typeStr = "double"
	case protoreflect.FloatKind:
		typeStr = "float"
	case protoreflect.Int64Kind:
		typeStr = "int64"
	case protoreflect.Uint64Kind:
		typeStr = "uint64"
	case protoreflect.Int32Kind:
		typeStr = "int32"
	case protoreflect.Fixed64Kind:
		typeStr = "fixed64"
	case protoreflect.Fixed32Kind:
		typeStr = "fixed32"
	case protoreflect.BoolKind:
		typeStr = "bool"
	case protoreflect.StringKind:
		typeStr = "string"
	case protoreflect.BytesKind:
		typeStr = "bytes"
	case protoreflect.Uint32Kind:
		typeStr = "uint32"
	case protoreflect.Sfixed32Kind:
		typeStr = "sfixed32"
	case protoreflect.Sfixed64Kind:
		typeStr = "sfixed64"
	case protoreflect.Sint32Kind:
		typeStr = "sint32"
	case protoreflect.Sint64Kind:
		typeStr = "sint64"
	case protoreflect.EnumKind:
		typeStr = "enum"
	case protoreflect.MessageKind:
		typeStr = "message"
	default:
		typeStr = "unknown"
	}
	return typeStr
}

// // GetProtoInfo 获取解析信息（结构化）
// func (p *ProtoParser) GetProtoInfo() map[string]interface{} {
// 	result := make(map[string]interface{})

// 	for fileName, fileDesc := range p.Files {
// 		fileInfo := map[string]interface{}{
// 			"syntax":   fileDesc.GetSyntax(),
// 			"package":  fileDesc.GetPackage(),
// 			"imports":  []string{},
// 			"messages": []map[string]interface{}{},
// 			"enums":    []map[string]interface{}{},
// 			"services": []map[string]interface{}{},
// 		}

// 		// 添加导入信息
// 		for _, dep := range fileDesc.GetDependencies() {
// 			fileInfo["imports"] = append(fileInfo["imports"].([]string), dep.GetName())
// 		}

// 		// 添加消息信息
// 		for _, msg := range fileDesc.GetMessageTypes() {
// 			fileInfo["messages"] = append(fileInfo["messages"].([]map[string]interface{}), p.getMessageInfo(msg))
// 		}

// 		// 添加枚举信息
// 		for _, enum := range fileDesc.GetEnumTypes() {
// 			fileInfo["enums"] = append(fileInfo["enums"].([]map[string]interface{}), p.getEnumInfo(enum))
// 		}

// 		// 添加服务信息
// 		for _, service := range fileDesc.GetServices() {
// 			fileInfo["services"] = append(fileInfo["services"].([]map[string]interface{}), p.getServiceInfo(service))
// 		}

// 		result[fileName] = fileInfo
// 	}

// 	return result
// }

// // getMessageInfo 获取消息信息
// func (p *ProtoParser) getMessageInfo(msg *desc.MessageDescriptor) map[string]interface{} {
// 	msgInfo := map[string]interface{}{
// 		"name":            msg.GetName(),
// 		"fields":          []map[string]interface{}{},
// 		"nested_messages": []map[string]interface{}{},
// 		"enums":           []map[string]interface{}{},
// 	}

// 	for _, field := range msg.GetFields() {
// 		fieldInfo := map[string]interface{}{
// 			"name":     field.GetName(),
// 			"type":     p.getFieldType(field),
// 			"number":   field.GetNumber(),
// 			"optional": field.IsOptional(),
// 			"repeated": field.IsRepeated(),
// 			"required": field.IsRequired(),
// 		}
// 		msgInfo["fields"] = append(msgInfo["fields"].([]map[string]interface{}), fieldInfo)
// 	}

// 	for _, nestedMsg := range msg.GetNestedMessageTypes() {
// 		msgInfo["nested_messages"] = append(msgInfo["nested_messages"].([]map[string]interface{}), p.getMessageInfo(nestedMsg))
// 	}

// 	for _, enum := range msg.GetEnumTypes() {
// 		msgInfo["enums"] = append(msgInfo["enums"].([]map[string]interface{}), p.getEnumInfo(enum))
// 	}

// 	return msgInfo
// }

// // getEnumInfo 获取枚举信息
// func (p *ProtoParser) getEnumInfo(enum *desc.EnumDescriptor) map[string]interface{} {
// 	enumInfo := map[string]interface{}{
// 		"name":   enum.GetName(),
// 		"values": []map[string]interface{}{},
// 	}

// 	for _, value := range enum.GetValues() {
// 		valueInfo := map[string]interface{}{
// 			"name":  value.GetName(),
// 			"value": value.GetNumber(),
// 		}
// 		enumInfo["values"] = append(enumInfo["values"].([]map[string]interface{}), valueInfo)
// 	}

// 	return enumInfo
// }

// // getServiceInfo 获取服务信息
// func (p *ProtoParser) getServiceInfo(service *desc.ServiceDescriptor) map[string]interface{} {
// 	serviceInfo := map[string]interface{}{
// 		"name":    service.GetName(),
// 		"methods": []map[string]interface{}{},
// 	}

// 	for _, method := range service.GetMethods() {
// 		methodInfo := map[string]interface{}{
// 			"name":        method.GetName(),
// 			"input_type":  method.GetInputType().GetFullyQualifiedName(),
// 			"output_type": method.GetOutputType().GetFullyQualifiedName(),
// 		}
// 		serviceInfo["methods"] = append(serviceInfo["methods"].([]map[string]interface{}), methodInfo)
// 	}

// 	return serviceInfo
// }
