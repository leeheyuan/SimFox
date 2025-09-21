package utils

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"encoding/json"

	"github.com/go-redis/redis"
	"github.com/magiconair/properties/assert"
)

func TestUT(t *testing.T) {
	t.SkipNow()
	{
		var arr []int
		for i := 0; i < 4; i++ {
			arr = append(arr, i%2)
		}
		n := TryRemoveByValue(&arr, 1)
		assert.Equal(t, arr, []int{0, 0})
		assert.Equal(t, n, 2)
	}

	{
		var arr = make([]int, 4)
		for i := 0; i < 4; i++ {
			arr[i] = 1
		}
		n := TryRemoveByValue(&arr, 1)
		assert.Equal(t, arr, []int{})
		assert.Equal(t, n, 4)
	}

	{
		var arr = []int{}
		n := TryRemoveByValue(&arr, 1)
		assert.Equal(t, arr, []int{})
		assert.Equal(t, n, 0)
	}

	{
		var arr = make([]int, 4)
		for i := 0; i < 4; i++ {
			arr[i] = i
		}

		ok := UniqueAppend(&arr, 0)
		assert.Equal(t, arr, []int{0, 1, 2, 3})
		assert.Equal(t, ok, false)

		ok = UniqueAppend(&arr, 4)
		assert.Equal(t, arr, []int{0, 1, 2, 3, 4})
		fmt.Println(arr)
		assert.Equal(t, ok, true)
	}
}

func TestComputer(t *testing.T) {

	fmt.Println(GetCPUFrequen())
	fmt.Println(GetCpuPercent())
	fmt.Println(GetMemPercent())
	fmt.Println(GetDiskPercent())
	fmt.Println(GetPhysicalCorenum())

	// 查询当前内存状态
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 查询内存剩余大小
	fmt.Printf("Memory remaining: %v bytes\n", m.Sys-m.HeapInuse)
}

func TestRound(t *testing.T) {
	t.SkipNow()
	assert.Equal(t, Round(0), 0)
	assert.Equal(t, Round(100), 100)
	assert.Equal(t, Round(11.4444), 11)
	assert.Equal(t, Round(22.666), 23)
	assert.Equal(t, Round(-22.666), -23)
	assert.Equal(t, Round(-11.111), -11)
}

func TestRand(t *testing.T) {
	t.SkipNow()
	rd1 := rand.New(rand.NewSource(9999))
	rd2 := rand.New(rand.NewSource(9999))

	for i := 0; i < 10; i++ {
		assert.Equal(t, RandomNum(0, 10000, rd1), RandomNum(0, 10000, rd2))
	}
}

func TestExists(t *testing.T) {
	t.SkipNow()
	var arr = []int{1, 2, 3, 4, 5}
	assert.Equal(t, Exists(arr, 1), true)
	assert.Equal(t, Exists(arr, 7), false)
}

func TestUpperBoundIdx(t *testing.T) {
	t.SkipNow()

	var arr = []int{1, 2, 4, 7, 7, 7, 10, 1000}
	ret := UpperBoundIdx(arr, 3)
	assert.Equal(t, ret, 2)

	ret = UpperBoundIdx(arr, 1)
	assert.Equal(t, ret, 1)

	ret = UpperBoundIdx(arr, 0)
	assert.Equal(t, ret, 0)

	ret = UpperBoundIdx(arr, 6)
	assert.Equal(t, ret, 3)

	ret = UpperBoundIdx(arr, 1000000)
	// assert.Equal(t, ret, 3)
	fmt.Println("upper:", ret)
}

type ConfigShuffle struct {
	Head    [9]int
	Next    [9]int
	Invalid [9]int

	BaidaMul [5]int
}

func TestCompileJSONWithComment(t *testing.T) {
	t.SkipNow()
	var a ConfigShuffle
	src := `{
    "Head": [
        100,/*物品1*/
        100,/*物品2*/
        80, /*物品3*/
        80, /*物品4*/
        180, /*物品5*/
        180, /*物品6*/
        120, /*物品7*/
        120, /*物品8*/
        40 /*物品9*/
    ],
    "Next": [
        160, /*正则*/
        160, //fdfd
        120,
        120,
        250,
        250,
        150,
        150,
        80
    ],
    "Invalid": [
        60,
        60,
        80,
        80,
        230,
        240,
        120,
        120,
        10
    ],
    "BaidaMul": [
        500,
        300,
        200,
        100,
        50
    ]
}`
	bb := CompileJSONWithComment([]byte(src))
	fmt.Println(string(bb))
	err := json.Unmarshal(bb, &a)
	if err != nil {
		t.Error(err)
	}
	// fmt.Println(a)
}

func TestHole(t *testing.T) {
	t.SkipNow()
	for i := 0; i < 100; i++ {
		ret := HoleArrayWithEnd(20, 3)
		// fmt.Println(ret)
		var sum int
		for _, b := range ret {
			if b {
				sum++
			}
		}
		assert.Equal(t, sum, 3)
		assert.Equal(t, ret[19], true)
	}
}

func TestHole2(t *testing.T) {
	t.SkipNow()
	// rand.Seed(time.Now().Unix())
	ret := HoleArrayWithEnd(10, 10)
	fmt.Println(ret)
}

func TestRandomFetch(t *testing.T) {
	t.SkipNow()
	// for i := 0; i < 100; i++ {
	arr := []int{1, 2, 3, 4, 5, 6, 7}
	ret := RandomFetch(arr, 0)
	fmt.Println(ret)
	// assert.Equal(t, ret[0] != ret[1], true)
	// }
}

func TestRandomFromPrioArr(t *testing.T) {
	t.SkipNow()
	for i := 0; i < 10; i++ {
		choice := RandomFromPrioArr([]int{5, 0, 5, 0, 5})
		fmt.Println(choice)
		assert.Equal(t, choice%2, 0)
	}
}

func TestSplitIntWithDLimit(t *testing.T) {
	t.SkipNow()
	arr, _ := SplitIntWithDLimit(370000, 1000, 120, 600)

	var min, max int
	min = arr[0]
	max = arr[0]
	for _, v := range arr {
		if v < min {
			min = v
		}

		if v > max {
			max = v
		}
	}
	fmt.Println(arr)
	fmt.Println(min, max)
}

func TestLshift(t *testing.T) {
	t.SkipNow()
	arr := []int{1, 2, 3, 4, 5}
	Lshift(arr)
	fmt.Println(arr)
	assert.Equal(t, arr, []int{2, 3, 4, 5, 1})
}

func TestTime2YMD(t *testing.T) {
	t.SkipNow()
	fmt.Println(Time2YMD(time.Now()))
}

func TestHoleArrayWithInterval(t *testing.T) {
	t.SkipNow()
	DEBUG = true
	holes := HoleArrayWithInterval(100, 3, 1, 80)
	fmt.Println(holes)
}

func TestMidnightTime(t *testing.T) {
	t.SkipNow()
	fmt.Println(MidnightTime())
}

func TestBase64(t *testing.T) {
	t.SkipNow()
	s := IntToBase64(12345)
	fmt.Println(s)
	i, _ := Base64ToInt(s)
	fmt.Println("i", i)
}

func TestLastSundayMidnightTimeBy(t *testing.T) {
	time := LastSundayMidnightTimeBy(time.Now().AddDate(0, 0, -5))
	fmt.Println("t", time)
}

func TestProgressStart(t *testing.T) {
	var cmd *exec.Cmd
	cmd = exec.Command("E:/item/package/UnityProject.exe", "-PixelStreamingURL=ws://127.0.0.1:89", "-AudioMixer", "-RenderOffScreen")
	// 创建子进程命令
	// 启动子进程
	err := cmd.Start()
	if err != nil {
		fmt.Println("Failed to start child process:", err)
		return
	}
	// 获取子进程ID
	pid := cmd.Process.Pid

	fmt.Println("Child process ID:", pid)
}

func TestGoroutine(t *testing.T) {

	ch := make(chan int, 3)

	go func() {
		time.Sleep(6 * time.Second)
		ch <- 10
	}()

	go func() {
		for {
			select {
			case data := <-ch:
				fmt.Println("Received data:", data)
			case <-time.After(3 * time.Second):
				fmt.Println("Timeout occurred")
				return
			}
		}

	}()

	for {
		time.Sleep(6 * time.Second)
	}

}

func WriteAuthorizationCode() {
	// 获取用户主目录路径
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("无法获取用户主目录路径:", err)
		return
	}
	// 构建隐私路径
	privateDir := filepath.Join(homeDir, ".systemdata")
	// 创建隐私路径目录
	err = os.MkdirAll(privateDir, 0700)
	if err != nil {
		fmt.Println("无法创建隐私路径目录:", err)
		return
	}
	// 写入文件到隐私路径
	filePath := filepath.Join(privateDir, "myfile.txt")
	if !FileExists(filePath) {
		data := []byte("这是要写入的文件内容")
		err = os.WriteFile(filePath, data, 0600)
		if err != nil {
			fmt.Println("无法写入文件:", err)
			return
		}
		fmt.Println("文件已写入到隐私路径:", filePath)
	} else {
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Println("无法写入文件:", err)
			return
		}
		fmt.Println("文件已写入到隐私路径:", ToString(data))
	}
}

func TestRegistrationAuthorizationCode(t *testing.T) {
	WriteAuthorizationCode()
}

func TestNet(t *testing.T) {
	// 连接到服务器
	conn, err := net.Dial("tcp", "example.com:8080")
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()
	// 发送数据
	message := "Hello, Server!"
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error sending message:", err)
		return
	}
	// 接收服务器响应
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error receiving response:", err)
		return
	}
	response := string(buffer[:n])
	fmt.Println("Response from server:", response)
}

func TestTime(t *testing.T) {

	name := GetApplicationName("E:\\EvidSimu3.0.1_KunMing\\Build\\海埂大坝_0705\\EivdSim3.1.10.exe")
	log.Println("nowtime:", name)
	sysTime := time.Now()
	log.Println("nowtime:", sysTime.GoString())
	duration, err := time.ParseDuration("1h2m3s")
	if err == nil {
		log.Println("time:", sysTime.Add(duration).Unix())
		time1 := time.Unix(sysTime.Add(duration).Unix(), 0)
		log.Println("time1:", time1.GoString())
	}

}

func TestNumberSlice(t *testing.T) {

	n := 15
	num1, num2 := DecomposeInteger(n)

	if num1 != 0 {
		fmt.Printf("整数 %d 分解为 %d 和 %d，它们的乘积为 %d\n", n, num1, num2, n)
	} else {
		fmt.Println("无法找到符合条件的分解整数的组合")
	}
}

func TestRedis(t *testing.T) {
	rediscon := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, perr := rediscon.Ping().Result()
	if perr != nil {
		fmt.Println("Failed to connect to Redis:", perr)
		return
	}
	//rediscon.Del("user:1")
	isexist, err1 := rediscon.HExists("user:1", "account").Result()
	if err1 != nil || !isexist {
		rediscon.HMSet("user:1", map[string]interface{}{"account": "test", "password": "123456"})
	} else {
		result, err := rediscon.HMGet("user:1", "account", "password").Result()
		if err == nil {
			v1, r1 := result[0].(string)
			if r1 {
				fmt.Println(v1)
			}
		}
	}

	/*result, err := rediscon.Do("SAdd", "AuthorizationCode", "123", "123", "456").Result()
	if err != nil {

		fmt.Println("fail")
	} else {
		fmt.Println("result")
		fmt.Println(result)
	}*/

	result, err := rediscon.Do("SMEMBERS", "AuthorizationCode").Result()

	{

		result, err := rediscon.SIsMember("AuthorizationCode", "f4a89924-6552-4a70-acba-6417b15a72fa").Result()
		if err != nil {
			// 处理错误
			fmt.Println("fail")
		} else {
			// 处理结果
			fmt.Println("result")
			fmt.Println(result)
		}

	}

	//result, err := rediscon.Do("SISMEMBER", "AuthorizationCode", "1123").Result()

	if err != nil {
		// 处理错误
		fmt.Println("fail")
	} else {
		// 处理结果
		fmt.Println("result")
		fmt.Println(result)
	}

}
