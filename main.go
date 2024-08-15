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
var nextPageButton selenium.WebElement

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

func findNextPageButton() (find bool) {
	waitCount := 0
	var err error
	for {
		nextPageButton, err = wb.FindElement(selenium.ByClassName, "ant-pagination-next")

		if err != nil {
			if waitCount > 5 {
				log.Println("没有找到下一页按钮")
				find = false
				break
			}
			log.Println("没有找到下一页按钮,继续等待...")
			time.Sleep(5 * time.Second)
			waitCount++
			continue
		} else {
			attribute, _ := nextPageButton.GetAttribute("aria-disabled")
			if attribute == "true" {
				log.Println("下一页按钮被禁用,当前是最后一页")
				find = false
				break
			}
			find = true
			break
		}
	}
	return find
}

// CheckClassProcessDone 获取当前章节的观看进度
func CheckClassProcessDone(i, all int) (bool, error) {

	var processElement selenium.WebElement
	var err error
	count := 0
	for {
		if count > 5 {
			return false, err
		}
		switchFrame("top")
		processElement, err = wb.FindElement(selenium.ByClassName, "vj-6f86217f")
		if err != nil {
			log.Println("没有找到进度条,5秒后重试")
			time.Sleep(5 * time.Second)
			count++
			continue
		}
		text, _ := processElement.Text()
		processText := strings.ReplaceAll(text, "%", "")
		process, _ := strconv.Atoi(processText)
		title, _ := wb.Title()
		fmt.Printf("\r%d / %d ==《 %s 》== 观看进度: %s %% ", i, all, title, processText)
		//这里判断视频是否正在播放时防止出现alert
		return process == 100, nil
	}
}

// CheckVideoPlaying 检查视频是否正在播放
func CheckVideoPlaying() bool {
	switchFrame("video")
	_, err := wb.FindElement(selenium.ByClassName, "vjs-playing")
	return err == nil
}

// CheckTextReading 检查是否观看电子书
func CheckTextReading() bool {
	switchFrame("top")
	_, err := wb.FindElement(selenium.ByID, "iframetext")
	return err == nil
}

func HandleVideo(i, all int, link string) {
	wb.Get(link)
	time.Sleep(5 * time.Second)
	element, err := wb.FindElement(selenium.ByClassName, "ant-btn-primary")
	if err != nil {
		log.Println("没有找到开始学习或继续学习按钮,观看下一课程！")
		return
	}
	element.Click()
	time.Sleep(5 * time.Second)
	for {
		done, err2 := CheckClassProcessDone(i, all)
		if err2 != nil {
			break
		}
		// 进度到达100%而且视频不再播放才寻找下一章节
		if done {
			fmt.Printf("\n")
			log.Println("当前视频播放完成，正在判断有没有下一节")
			switchFrame("top")
			findElement, err22 := wb.FindElement(selenium.ByClassName, "vj-99c3bcc7")
			if err22 != nil {
				log.Println("没有找到更多章节,继续观看下一链接")
				break
			}
			text, _ := findElement.Text()
			if text == "下一节" {
				findElement.Click()
				time.Sleep(3 * time.Second)
				//判断弹窗
				but, err3 := wb.FindElement(selenium.ByClassName, "ant-btn-primary")
				if err3 == nil {
					fmt.Println("找到确定按钮")
					but.Click()
				}

				continue
			} else {
				log.Println("不是按钮不是下一节,继续观看下一链接")
				break
			}
		} else {
			playing := CheckVideoPlaying()
			if playing {
				time.Sleep(10 * time.Second)
				continue
			}
			reading := CheckTextReading()
			if reading {
				time.Sleep(10 * time.Second)
				continue
			}
			//尝试点击播放按钮
			switchFrame("video")
			playButton, err3 := wb.FindElement(selenium.ByClassName, "vjs-button-icon")
			fmt.Printf("\n")
			if err3 == nil {
				playButton.Click()
				log.Println("点击了播放按钮")
				time.Sleep(2 * time.Second)
			} else {
				log.Println("没有找到播放按钮,继续观看下一链接")
				break
			}
		}

	}
	return

}

func HandlePageLinks() []string {
	links := make([]string, 0)
	elements, err := wb.FindElements(selenium.ByClassName, "vj-a1d5dd58")
	if err != nil {
		log.Println("没有找到课程列表")
		return links
	}
	log.Println("本页共发现", len(elements), "个课程")
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
	log.Println("本页共收集", len(links), "个课程")
	return links
}

func CollectionLinks(url string) []string {
	links := make([]string, 0)
	if url != "" {
		wb.Get(url)
	}
	log.Println("等待页面加载完成,您可以手动切换至更多分页以加快视频收集速度")
	time.Sleep(10 * time.Second)

	for {
		_, err := wb.FindElement(selenium.ByClassName, "vj-a1d5dd58")
		if err != nil {
			log.Println("等待页面加载中...")
			time.Sleep(5 * time.Second)
			continue
		} else {
			break
		}
	}

	links = append(links, HandlePageLinks()...)

	for findNextPageButton() {
		nextPageButton.Click()
		time.Sleep(5 * time.Second)
		tLinks := HandlePageLinks()
		links = append(links, tLinks...)
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
		//DebuggerAddr: "127.0.0.1:9222",
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
	log.Println("链接 webDriver 成功")
	return true
}

func main() {

	success := InitWebDriver()
	if success == false {
		log.Println("初始化WebDriver失败")
		return
	}
	log.Println("初始化WebDriver 成功")
	var url string
	url = "https://qy.51vj.cn/app/home/school/course/1/0?appid=1003&corpid=wp58yYCQAAx65D52VX68yo_9ZU37eTgQ"

	links := CollectionLinks(url)
	for i, link := range links {
		fmt.Println("第", i+1, "个课程：", link)
		HandleVideo(i+1, len(links), link)
	}
	log.Println(len(links), "个课程播放完成！")
	service.Stop()
	wb.Quit()
	select {}

}
