package main

import (
	"fmt"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"log"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var service *selenium.Service
var wb selenium.WebDriver

func switchFrame(op string) {
	switch op {
	case "video":
		_ = wb.SwitchFrame(nil)
		elementVideo, _ := wb.FindElement(selenium.ByID, "vj-weschool-video-iframe")
		_ = wb.SwitchFrame(elementVideo)
	case "top":
		_ = wb.SwitchFrame(nil)
	}
}

func GetCurrentVideoProcess() int {
	switchFrame("top")
	var processElement selenium.WebElement
	var err error
	for {
		processElement, err = wb.FindElement(selenium.ByClassName, "vj-6f86217f")
		if err != nil {
			log.Println("没有找到进度条,5秒后重试")
			time.Sleep(5 * time.Second)
			continue
		}
		text, _ := processElement.Text()
		processText := strings.ReplaceAll(text, "%", "")
		process, _ := strconv.Atoi(processText)
		return process
	}
}

func HandleVideo(i, all int, link string) {
	wb.Get(link)
	time.Sleep(5 * time.Second)
	element, err := wb.FindElement(selenium.ByClassName, "ant-btn-primary")
	if err != nil {
		log.Println("没有找到观看按钮,继续观看下一课程！")
		return
	}
	element.Click()
	time.Sleep(5 * time.Second)
	for {
		videoInitProcess := GetCurrentVideoProcess()
		time.Sleep(5 * time.Second)
		title, _ := wb.Title()
		if videoInitProcess == 100 {
			log.Println("当前章节观看完成,正在判断是否存在下一章节")
			switchFrame("top")
			findElement, err22 := wb.FindElement(selenium.ByClassName, "vj-99c3bcc7")
			if err22 != nil {
				log.Println("没有找到更多章节,继续观看下一链接")
				return
			}
			text, _ := findElement.Text()
			if text == "下一节" {
				findElement.Click()
				continue
			} else {
				log.Println("没有找到下一节按钮,继续观看下一链接")
				return
			}
		} else {
			log.Println(i, "/", all, "->", title, " == >观看进度:", videoInitProcess, "%")
		}
		switchFrame("video")

		_, err1 := wb.FindElement(selenium.ByClassName, "vjs-playing")
		if err1 == nil {
			log.Println("视频正在播放中...10秒后继续!")
			time.Sleep(10 * time.Second)
			continue
		} else {
			findElement2, err2 := wb.FindElement(selenium.ByClassName, "vjs-button-icon")
			if err2 == nil {
				findElement2.Click()
				log.Println("点击了播放按钮")
			} else {
				log.Println("视频播放完成")
				break
			}
		}
	}
	return

}

func CollectionLinks() []string {
	links := make([]string, 0)
	wb.Get("https://qy.51vj.cn/app/home/school/course/1/0?appid=1003&corpid=wp58yYCQAAx65D52VX68yo_9ZU37eTgQ")
	for {
		_, err := wb.FindElement(selenium.ByClassName, "ant-pagination")
		if err != nil {
			log.Println("没有找到分页,继续等待5秒")
			time.Sleep(5 * time.Second)
			continue
		} else {
			break
		}
	}
	for {
		_, err := wb.FindElement(selenium.ByClassName, "ant-pagination")
		if err == nil {
			log.Println("找到分页,请手动选择120条/页")
			time.Sleep(5 * time.Second)
			continue
		} else {
			log.Println("切换完成,程序继续")
			break
		}
	}
	time.Sleep(5 * time.Second)

	elements, _ := wb.FindElements(selenium.ByClassName, "vj-a1d5dd58")
	log.Println("共发现", len(elements), "个课程")
	for _, element := range elements {
		element1, err1 := element.FindElement(selenium.ByClassName, "vj-74b8d6e3")
		if err1 != nil {
			continue
		}
		element2, err2 := element1.FindElement(selenium.ByClassName, "vj-03ca486b")
		if err2 != nil {
			continue
		}
		text, _ := element2.Text()
		if text != "已完成" && text != "" {
			link, _ := element.GetAttribute("href")
			links = append(links, link)
		}
	}
	log.Println("共", len(links), "个课程需要观看")
	return links
}

func InitWebDriver() bool {
	var opts []selenium.ServiceOption
	sysType := runtime.GOOS
	driverName := "chromedriver"
	switch sysType {
	case "windows":
		driverName = "chromedriver.exe"
	case "linux":
		driverName = "chromedriver_linux"
	case "darwin":
		driverName = "chromedriver_macos"
	}
	log.Println("系统类型:", sysType, "浏览器驱动:", driverName)
	var err error
	service, err = selenium.NewChromeDriverService("./drivers/"+driverName, 9515, opts...)
	if nil != err {
		log.Println("启动 chromedriver 失败，err: ", err.Error())
		return false
	}

	//链接本地的浏览器 chrome
	caps := selenium.Capabilities{
		"browserName": "chrome",
	}

	chromeCaps := chrome.Capabilities{
		Path: "",
		Args: []string{
			"--user-agent=Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36",
			"--remote-allow-origins=*",
		},
	}
	//以上是设置浏览器参数
	caps.AddChrome(chromeCaps)

	// 调起chrome浏览器
	wb, err = selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", 9515))
	if err != nil {
		log.Println("链接 webDriver 失败", err.Error())
		return false
	}
	return true
}

func main() {

	success := InitWebDriver()
	if success == false {
		log.Println("初始化WebDriver失败")
		return
	}

	links := CollectionLinks()
	for i, link := range links {
		HandleVideo(i+1, len(links), link)
	}
	log.Println(len(links), "个课程播放完成！")
	service.Stop()
	wb.Quit()
	select {}

}
