package main

import (
	"flag"
	"fmt"
	"github.com/zsinx/go2rpcx/src"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path"
	"strings"
	"text/template"
	"unicode"
)

const streamFlag = "_stream"
const pingpang = "_pingpang"

var filePath string

// var dir string
var target string

func init() {
	flag.StringVar(&filePath, "f", "", "source file path")
	flag.StringVar(&target, "t", "rpc", "rpc file target path")
	flag.Usage = usage
}

func main() {
	flag.Parse()
	if filePath == "" {
		flag.Usage()
		return
	}
	start(filePath)
}

func start(interfacePath string) {
	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, interfacePath, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	microService := src.MicroService{
		FileName: interfacePath,
		PackageName: target,
	}
	imports := []string{}
	messages := []src.Message{}
	ast.Inspect(f, func(node ast.Node) bool {
		if node == nil {
			return true
		}
		if importSpec, ok := node.(*ast.ImportSpec); ok {
			if importSpec.Path.Value == "\"time\"" {
				microService.ImportTime = true
			} else {
				importpath := importSpec.Path.Value
				imports = append(imports, importpath)
			}
		}

		if typeSpecNode, ok := node.(*ast.TypeSpec); ok {
			// 处理接口
			if interfaceNode, f := typeSpecNode.Type.(*ast.InterfaceType); f {
				fmt.Println("接口名称：", typeSpecNode.Name.Name)
				serviceFunctions := interfaceParser(interfaceNode)
				service := src.Service{}
				service.Name = typeSpecNode.Name.Name
				service.PackageName = genPackageName(typeSpecNode.Name.Name)
				service.ServiceFunctions = serviceFunctions
				microService.Service = service
				// spew.Dump(serviceFunctions)
			}
			// 处理结构体
			if structNode, f := typeSpecNode.Type.(*ast.StructType); f {
				structName := typeSpecNode.Name.Name
				log.Println("struct名称：", structName)
				messageFields := structParser(structName, structNode)
				message := src.Message{}
				message.Name = structName
				message.MessageFields = messageFields
				messages = append(messages, message)
				// spew.Dump(messageFields)
			}

		}
		return true
	})
	microService.Messages = messages
	microService.Imports = imports

	targetFileName := strings.Replace(path.Base(interfacePath), ".go", ".rpc.go", -1)
	saveToFile(microService, fmt.Sprintf("%v/%v", target, targetFileName))
}

func genPackageName(s string) string {
	if unicode.IsUpper([]rune(s)[0]) {
		return strings.ToLower(string(s[0])) + string(s[1:])
	}
	return s
}

/*
rpc 文件模板
*/
func saveToFile(microService src.MicroService, rpcPath string) {
	t, err := template.New("RpcTemplate").Parse(src.RpcTemplate)
	if err != nil {
		log.Panic(err)
	}
	os.MkdirAll(path.Dir(rpcPath), 0755)
	os.Remove(rpcPath)
	f, err := os.OpenFile(rpcPath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		log.Panic(err)
	}

	err = t.Execute(f, microService)
	if err != nil {
		log.Panic(err)
	}
}

/*
解析结构体
*/
func structParser(structName string, structNode *ast.StructType) []src.MessageField {
	messageFields := []src.MessageField{}
	for i, field := range structNode.Fields.List {
		messageField := src.MessageField{}
		messageField.Index = i + 1
		// 如果是类似name,address string 这样的定义则报错
		if len(field.Names) != 1 {
			log.Fatalf("struct %v error,the field can't define like 'name,address string'", structName)
		}
		messageField.FieldName = field.Names[0].Name

		// 基本类型处理
		if fieldType, ok := field.Type.(*ast.Ident); ok {
			messageField.FieldType = fieldType.Name
		}

		// map类型处理
		if fieldType, ok := field.Type.(*ast.MapType); ok {
			key, value := "", ""
			if keyType, ok := fieldType.Key.(*ast.Ident); ok {
				key = keyType.Name
			}
			if valueType, ok := fieldType.Value.(*ast.Ident); ok {
				value = valueType.Name
			}
			messageField.FieldType = fmt.Sprintf("map<%v,%v>", key, value)
		}

		// 处理引用类型
		if fieldType, ok := field.Type.(*ast.SelectorExpr); ok {
			if p, ok := fieldType.X.(*ast.Ident); ok && p.Name == "time" {
				messageField.FieldType = "time.Time"
			} else {
				messageField.FieldType = fieldType.Sel.Name
			}
		}
		// 处理参数是数组的情况
		if fieldType, ok := field.Type.(*ast.ArrayType); ok {
			if fieldTypeElt, ok := fieldType.Elt.(*ast.Ident); ok {
				messageField.FieldType = "[]" + fieldTypeElt.Name
			}
		}
		// 获取标签
		if field.Tag != nil {
			messageField.FieldTag = field.Tag.Value
		}
		// 获取注释
		if field.Doc != nil {
			messageField.Comment = field.Doc.List[0].Text
		} else if field.Comment != nil {
			messageField.Comment = field.Comment.List[0].Text
		}
		messageFields = append(messageFields, messageField)
	}
	return messageFields
}

/*
解析接口代码
*/
func interfaceParser(interfaceNode *ast.InterfaceType) []src.ServiceFunction {
	serviceFunctions := []src.ServiceFunction{}
	// 解析方法列表
	for _, function := range interfaceNode.Methods.List {

		serviceFunction := src.ServiceFunction{}
		if function.Doc != nil {
			serviceFunction.Comment = function.Doc.List[0].Text
		} else if function.Comment != nil {
			serviceFunction.Comment = function.Comment.List[0].Text
		}
		// 获取方法名称
		if len(function.Names) != 1 {
			log.Fatal("parser function error")
		}

		functionName := function.Names[0].Name
		if strings.HasSuffix(functionName, streamFlag) {
			serviceFunction.Name = strings.Replace(functionName, streamFlag, "", -1)
			serviceFunction.Stream = true
		} else if strings.HasSuffix(functionName, pingpang) {
			serviceFunction.Name = strings.Replace(functionName, pingpang, "", -1)
			serviceFunction.PingPong = true
		} else {
			serviceFunction.Name = functionName
		}

		// 解析方法
		if funcBody, ok := function.Type.(*ast.FuncType); ok {
			// 解析参数列表
			for i, param := range funcBody.Params.List {
				// 获取参数名称
				for _, paramName := range param.Names {
					log.Printf("function:%v index:%v paramName:%v ", functionName, i+1, paramName.Name)
				}
				// 获取参数类型
				if paramType, ok := param.Type.(*ast.Ident); ok {
					serviceFunction.ParamType = paramType.Name
					log.Printf("function:%v index:%v paramType:%v ", functionName, i+1, paramType.Name)
				}
			}
			// 解析返回值
			for i, result := range funcBody.Results.List {
				// 获取参数名称
				for _, resultName := range result.Names {
					log.Printf("function:%v index:%v resultName:%v ", functionName, i+1, resultName.Name)
				}
				// 获取参数类型
				if resultType, ok := result.Type.(*ast.Ident); ok {
					serviceFunction.ResultType = resultType.Name
					log.Printf("function:%v index:%v resultType:%v ", functionName, i+1, resultType.Name)
				}
			}
		}
		serviceFunctions = append(serviceFunctions, serviceFunction)
	}
	return serviceFunctions
}

func usage() {
	fmt.Fprintf(os.Stderr, `version: 1.0
Usage: go2rpcx [-f] [-t]

Options:
`)
	flag.PrintDefaults()
}
