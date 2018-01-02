package service

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"os"
	"utils"
)
/*
type Cmd struct {
    Path         string　　　//运行命令的路径，绝对路径或者相对路径
    Args         []string　　 // 命令参数
    Env          []string         //进程环境，如果环境为空，则使用当前进程的环境
    Dir          string　　　//指定command的工作目录，如果dir为空，则comman在调用进程所在当前目录中运行
    Stdin        io.Reader　　//标准输入，如果stdin是nil的话，进程从null device中读取（os.DevNull），stdin也可以时一个文件，否则的话则在运行过程中再开一个goroutine去
　　　　　　　　　　　　　//读取标准输入
    Stdout       io.Writer       //标准输出
    Stderr       io.Writer　　//错误输出，如果这两个（Stdout和Stderr）为空的话，则command运行时将响应的文件描述符连接到os.DevNull
    ExtraFiles   []*os.File 　　
    SysProcAttr  *syscall.SysProcAttr
    Process      *os.Process    //Process是底层进程，只启动一次
    ProcessState *os.ProcessState　　//ProcessState包含一个退出进程的信息，当进程调用Wait或者Run时便会产生该信息．
}

# test.sh
#!/bin/bash

for k in $( seq 1 10 )
do
   echo "Hello World $k"
   sleep 1
done
 */

func main() {
	command := "/bin/bash"
	params := []string{"-c", "sh /Users/zhangbaozhen/codespace/hhmedic/script/test.sh"}

	ExecCommandFile(command, params)
}

func ExecCommandFile(commandName string, params []string) (bool, []string) {
	var contentArray = make([]string, 0, 5)
	//contentArray = contentArray[0:0]
	cmd := exec.Command(commandName, params...)
	//显示运行的命令
	fmt.Printf("执行命令: %s\n", strings.Join(cmd.Args[1:], " "))
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error=>", err.Error())
		return false, contentArray
	}
	//cmd.Run() //开始指定命令并且等待他执行结束，如果命令能够成功执行完毕，则返回nil，否则的话边会产生错误
	cmd.Start() // Start开始执行c包含的命令，但并不会等待该命令完成即返回。Wait方法会返回命令的返回状态码并在命令返回后释放相关的资源。

	reader := bufio.NewReader(stdout)

	var index int
	//实时循环读取输出流中的一行内容
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		utils.Debugf("line=%s", line)
		index++
		contentArray = append(contentArray, line)
	}

	cmd.Wait()
	return true, contentArray
}

func ExecCommand(commandName string, params []string) (bool, []string) {
	var contentArray = make([]string, 0, 5)
	//contentArray = contentArray[0:0]
	cmd := exec.Command(commandName, params...)
	//显示运行的命令
	fmt.Printf("执行命令: Dir=%s \n Path=%s cmd=\n %s \n args= %s \n", cmd.Dir, cmd.Path, cmd, strings.Join(cmd.Args[1:], " "))
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error=>", err.Error())
		contentArray = append(contentArray, "error=> " + err.Error())
		return false, contentArray
	}
	cmd.Start() // Start开始执行c包含的命令，但并不会等待该命令完成即返回。Wait方法会返回命令的返回状态码并在命令返回后释放相关的资源。

	reader := bufio.NewReader(stdout)

	var index int
	//实时循环读取输出流中的一行内容
	for {
		line, err2 := reader.ReadString('\n')
		fmt.Printf("%d: %s \n", index, line)
		if err2!=nil && io.EOF==err2{
			contentArray = append(contentArray, "success.")
			break
		}else if err2 != nil {
			//contentArray = append(contentArray, "faild." + err2.Error())//不需要，Wait会catch到
			break
		}
		index++
		contentArray = append(contentArray, line)
	}

	err = cmd.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed with: ", err.Error())
		contentArray = append(contentArray, "failed with: " + err.Error())
		return false, contentArray
	}
	return true, contentArray
}