package utils

import (
	"archive/zip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	crand "crypto/rand"
	"path"
	"sync"

	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"

	"math"

	"math/rand"
	"os"
	"reflect"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/donnie4w/go-logger/logger"
	"github.com/gorilla/websocket"
	ps "github.com/mitchellh/go-ps"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

func MakeMsg(args ...interface{}) ([]byte, error) {
	return json.Marshal(args)
}

func OpenURL(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Run()
}

func LunchSumoSimulation(workdir string, path string) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("无法获取当前工作路径：", err)
		return
	}
	fmt.Println("当前工作路径：", dir)
	err = os.Chdir(workdir)
	if err != nil {
		fmt.Println("切换工作路径失败：", err)
		return
	}

	cmd := exec.Command("py", "sim_server.py", fmt.Sprintf("--config=%s", path))
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("执行出错:", err)
	}
	// 打印 Python 的输出结果
	fmt.Println(string(output))

}

func GetClientIP(r *http.Request) string {
	// 尝试从 X-Forwarded-For 头中获取
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For 可能包含多个 IP 地址，用逗号分隔，取第一个
		ips := strings.Split(ip, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 尝试从 X-Real-Ip 头中获取
	ip = r.Header.Get("X-Real-Ip")
	if ip != "" {
		return ip
	}

	// 最后从 RemoteAddr 中获取
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

func IsPortOpen(host string, port string) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), 2*time.Second)
	if err != nil {
		return false
	}

	defer conn.Close()
	return true
}

func FindInt16ArrIndex(arr []int16, target int16) int16 {
	for i, v := range arr {
		if v == target {
			return int16(i)
		}
	}
	return -1 // 如果没找到，返回 -1
}

func GetCPUFrequen() float64 {
	infos, err := cpu.Info()
	if err != nil {
		fmt.Println("Error:", err)
	}

	for i, info := range infos {
		fmt.Printf("CPU %d: %s @ %.2f MHz\n", i, info.ModelName, info.Mhz)
		return info.Mhz
	}
	return 0

}

func GetPhysicalCorenum() int {
	pcnt, err := cpu.Counts(false)
	if err != nil {
		return 0
	}
	return pcnt
}

func GetCpuPercent() float64 {
	percent, _ := cpu.Percent(time.Second, false)
	return percent[0]
}

func GetMemPercent() float64 {
	memInfo, _ := mem.VirtualMemory()
	return memInfo.UsedPercent
}

func GetDiskPercent() float64 {
	parts, _ := disk.Partitions(true)
	diskInfo, _ := disk.Usage(parts[0].Mountpoint)
	return diskInfo.UsedPercent
}

func GetSystemInfo() (float64, float64) {
	return GetCpuPercent(), GetMemPercent()
}

func DownloadFile(url string, fileName string, wg *sync.WaitGroup) string {
	defer func() {
		if wg != nil {
			wg.Done()
		}
	}()
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error downloading file:", err)
		return ""
	}
	defer resp.Body.Close()
	out, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return ""
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Println("Error copying file:", err)
		return ""
	}
	return fileName
}

func IsValidURL(inputURL string) bool {
	_, err := url.ParseRequestURI(inputURL)
	return err == nil
}

func RemoveDuplicateStrings(slice []string) []string {
	encountered := map[string]bool{}
	result := []string{}

	for v := range slice {
		if encountered[slice[v]] == false {
			encountered[slice[v]] = true
			result = append(result, slice[v])
		}
	}

	return result
}

func ServerReachable(address string) bool {
	conn, err := net.DialTimeout("tcp", address, 1*time.Second)
	if err != nil {
		fmt.Println("Server unreachable:", err)
		return false
	}
	defer conn.Close()
	fmt.Println("Server reachable")
	return true
}

func ProcessExistsEx(pid int) bool {
	if runtime.GOOS == "windows" {
		return WindowsisProcessRunning(pid)
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(os.Signal(syscall.Signal(0)))
	if err != nil {
		// 如果错误是 "ESRCH"，说明进程不存在
		if err == syscall.ESRCH {
			return false
		}
		// 如果错误是 "EPERM"，说明进程存在，但无权访问
		if err == syscall.EPERM {
			return true
		}
		// 其他错误
		return false
	}
	return true
}

func KillProcess(pid int) error {
	logger.Info("KillProcess:", pid)
	if runtime.GOOS == "linux" {
		// Linux 平台
		killCmd := exec.Command("kill", "-9", strconv.Itoa(pid))
		err := killCmd.Run()
		if err != nil {
			fmt.Println("Error killing process:", err)
			return err
		}
		// 等待进程退出并清理僵尸进程
		process, err := os.FindProcess(pid)
		if err != nil {
			fmt.Println("查找进程时出错:", err)
			return err
		}
		_, err = process.Wait()
		if err != nil {
			fmt.Println("清理僵尸进程时出错:", err)
			return err
		}
		fmt.Println("Process", pid, "and its child processes killed successfully")
	} else if runtime.GOOS == "windows" {
		// Windows 平台
		killCmd := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid))
		err := killCmd.Run()
		if err != nil {
			fmt.Println("Error killing process:", err)
			return err
		}

		fmt.Println("Process", pid, "and its child processes killed successfully")
	} else {
		fmt.Println("Unsupported operating system:", runtime.GOOS)
	}
	return nil
}

func ProcessExists(processName string) (bool, error) {
	processes, err := ps.Processes()
	if err != nil {
		return false, err
	}

	for _, p := range processes {
		if p.Executable() == processName {
			return true, nil
		}
	}

	return false, nil
}

func KillProcessByName(processName string) {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		// Windows 平台
		cmd = exec.Command("taskkill", "/F", "/T", "/IM", processName+".exe")
	} else {
		// Linux 平台
		pidCmd := exec.Command("pgrep", "-f", processName)
		pidOutput, err := pidCmd.Output()
		if err != nil {
			logger.Info("Error executing pgrep command:", err)
			return
		}
		pid := strings.TrimSpace(string(pidOutput))

		if pid == "" {
			logger.Info("Process", processName, "does not exist")
			return
		}

		cmd = exec.Command("pkill", "-P", pid)
	}

	err := cmd.Run()
	if err != nil {
		logger.Info("Error killing process:", err)
		return
	}

	logger.Info("Process", processName, "and its child processes killed successfully")
}

// 密钥需要32字节用于AES-256
var key = []byte("examplekey1234567890123456a8efw2")

// 加密函数
func Encrypt(plainText string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plainText))
	iv := ciphertext[:aes.BlockSize] // 创建初始向量

	if _, err := io.ReadFull(crand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plainText))

	// 返回 base64 编码后的加密字符串
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// 解密函数
func Decrypt(encryptedText string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}

func MakeMsgToString(args ...interface{}) string {
	ret, err := json.Marshal(args)
	if err == nil {
		return string(ret)
	}
	return ""
}

func MakeMsgFromArray(arr []interface{}) ([]byte, error) {
	return json.Marshal(arr)
}

func RandomNum(minNum int, maxNum int, rd *rand.Rand) int {
	if minNum == maxNum {
		return minNum
	}

	if rd != nil {
		return rd.Intn(maxNum-minNum+1) + minNum
	}
	return rand.Intn(maxNum-minNum+1) + minNum
}

// func RandomNum64(minNum int64, maxNum int64) int64 {
// 	return rand.Int63n(maxNum-minNum+1) + minNum
// }

func RandomFromArray(arr []int) int {
	idx := rand.Intn(len(arr))
	return arr[idx]
}

func Assert(b bool) {
	if !b {
		debug.PrintStack()
		logger.Error("断言失败-------------")
		panic("assert failed")
	}
}

func GetDayCount() int {
	return time.Now().YearDay()
}

func ConvertToIntEx(arg interface{}) int {
	var v int
	switch arg.(type) {
	case float64:
		v = int(arg.(float64))
		break
	case int64:
		v = int(arg.(int64))
		break
	case int:
		v = arg.(int)
		break
	case string:
		v, _ = strconv.Atoi(arg.(string))
		break
	}
	return v
}

func ConvertToInt64Ex(arg interface{}) (int64, bool) {
	var v int64
	var r bool = false
	switch arg.(type) {
	case float64:
		v = int64(arg.(float64))
		r = true
		break
	case float32:
		v = int64(arg.(float32))
		r = true
		break
	case int64:
		v = arg.(int64)
		r = true
		break
	case int:
		v = int64(arg.(int))
		r = true
		break
	case string:
		n, err := strconv.ParseInt(arg.(string), 10, 64)
		if err == nil {
			v = n
			r = true
		}
		break
	}
	return v, r
}

func StringExist(target string, str_array []string) bool {
	sort.Strings(str_array)
	index := sort.SearchStrings(str_array, target)
	//index的取值：[0,len(str_array)]
	if index < len(str_array) && str_array[index] == target { //需要注意此处的判断，先判断 &&左侧的条件，如果不满足则结束此处判断，不会再进行右侧的判断
		return true
	}
	return false

}

func ConvertToInt(args []interface{}, i int) (int, bool) {
	if len(args) <= i {
		return 0, false
	}

	if f, ok := args[i].(float64); ok {
		return int(f), true
	}

	return 0, false
}

func ConvertToInt64(args []interface{}, i int) (int64, bool) {
	if len(args) <= i {
		return 0, false
	}

	if f, ok := args[i].(float64); ok {
		return int64(f), true
	}

	if f, ok := args[i].(int64); ok {
		return int64(f), true
	}

	if f, ok := args[i].(string); ok {

		num, err := strconv.ParseInt(f, 10, 64)
		if err != nil {
			fmt.Println("无效的整数字符串:", err)
			return 0, false
		}
		return num, true
	}
	return 0, false
}

func ConvertToBool(args []interface{}, i int) (bool, bool) {
	if len(args) <= i {
		return false, false
	}

	if f, ok := args[i].(bool); ok {
		return f, true
	}

	return false, false
}

func ConvertToString(args []interface{}, i int) (string, bool) {
	if len(args) <= i {
		return "", false
	}

	if f, ok := args[i].(string); ok {
		return f, true
	}

	return "", false
}

func ConvertToArr(args []interface{}, i int) ([]interface{}, bool) {
	if len(args) <= i {
		return nil, false
	}

	if f, ok := args[i].([]interface{}); ok {
		return f, true
	}

	return nil, false
}

func ShuffleStringArr(arr []string) {
	for i := len(arr) - 1; i >= 1; i-- {
		var ridx = rand.Intn(i)
		arr[i], arr[ridx] = arr[ridx], arr[i]
	}
}

func ShuffleIntArr(arr []int) {
	for i := len(arr) - 1; i >= 1; i-- {
		var ridx = rand.Intn(i)
		arr[i], arr[ridx] = arr[ridx], arr[i]
	}
}

func ShufflebyteArr(arr []byte) {
	for i := len(arr) - 1; i >= 1; i-- {
		var ridx = rand.Intn(i)
		arr[i], arr[ridx] = arr[ridx], arr[i]
	}
}

func Sum(arr []int) int {
	var sum int
	for _, v := range arr {
		sum += v
	}

	return sum
}

func init() {
	rand.Seed(time.Now().Unix())
}

func TodayZero() time.Time {
	now := time.Now()
	t0 := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return t0
}

func Colorize(text string, status string) string {
	out := ""
	switch status {
	case "blue":
		out = "\033[32;1m" // Blue
	case "red":
		out = "\033[31;1m" // Red
	case "yellow":
		out = "\033[33;1m" // Yellow
	case "green":
		out = "\033[34;1m" // Green
	default:
		out = "\033[0m" // Default
	}
	return out + text + "\033[0m"
}

func IndexOf(arr []int, target int) int {
	for i, num := range arr {
		if num == target {
			return i
		}
	}
	return -1
}

func Exists(arr []int, target int) bool {
	return IndexOf(arr, target) != -1
}

func AllIs(arr []int, target int) bool {
	for _, v := range arr {
		if v != target {
			return false
		}
	}

	return true
}

func RemoveByIdx(arr *[]int, i int) {
	*arr = append((*arr)[:i], (*arr)[i+1:]...)
}

func TryRemoveByValue(arr *[]int, v int) int {
	var tail = -1
	for i := 0; i < len(*arr); i++ {
		findit := (*arr)[i] == v
		if tail == -1 {
			if findit {
				tail = i
			}
		} else {
			if !findit {
				(*arr)[tail] = (*arr)[i]
				tail++
			}
		}
	}

	if tail == -1 {
		return 0
	}

	ret := len(*arr) - tail

	*arr = (*arr)[:tail]

	return ret
}

func UniqueAppend(arr *[]int, v int) bool {
	if IndexOf(*arr, v) == -1 {
		*arr = append(*arr, v)
		return true
	}

	return false
}

func Round(v float64) int {
	i, f := math.Modf(v)
	if f < -0.5 {
		return int(i) - 1
	}

	if f < 0.5 {
		return int(i)
	}

	return int(i) + 1
}

// 1,1,1,1,2,2,2,3,3,3,
// 2,2,2,1,1,1,1,3,3,3
// 2,2,2,3,3,3,1,1,1,1
// 1,2,3,4,5,6,7
// 4,1,2,3,5,6,7
func MoveBlock(arr []int, len, dst int) {
	bak := make([]int, len)
	copy(bak, arr)

	end := len + dst
	for i := len; i < end; i++ {
		arr[i-len] = arr[i]
	}

	for i := 0; i < len; i++ {
		arr[i+dst] = bak[i]
	}
}

func UpperBoundIdx(arr []int, t int) int {
	lo := 0
	hi := len(arr) - 1

	for lo < hi {

		mid := (lo + hi) / 2

		v := arr[mid]
		if v <= t {
			lo = mid + 1
		} else {
			hi = mid
		}
	}

	return hi
}

func CompileJSONWithComment(src []byte) []byte {
	//pat := `(/\*.*\*/|//.*)`
	/*re, err := regexp.Compile(pat)
	if err != nil {
		log.Fatalln(err)
	}
	resutlStr := re.ReplaceAll(src, nil)*/

	return src
}

func UnmarshalJsonWithComment(content []byte, p interface{}) error {
	compileContent := CompileJSONWithComment(content)
	err := json.Unmarshal(compileContent, p)
	return err
}

func UnmarshalJsonFileWithComment(file string, p interface{}) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	compileContent := CompileJSONWithComment(content)
	err = json.Unmarshal(compileContent, p)
	return err
}

func HoleArray(size, n int) []bool {
	arr := make([]int, size)
	for i := 0; i < size; i++ {
		arr[i] = i
	}

	for i := 0; i < n; i++ {
		r := rand.Intn(size)
		arr[i], arr[r] = arr[r], arr[i]
	}

	// logger.Info("arr", arr)

	ret := make([]bool, size)
	for i := 0; i < n; i++ {
		ret[arr[i]] = true
	}
	return ret
}

func HoleArrayWithEnd(size, n int) []bool {
	ret := HoleArray(size-1, n-1)
	ret = append(ret, true)
	return ret
}

func RandomFetch(arr []int, n int) []int {
	end := MinInt(n, len(arr))
	for i := 0; i < end; i++ {
		r := rand.Intn(len(arr))
		arr[i], arr[r] = arr[r], arr[i]
	}
	return arr[:end]
}

func RandomFetchByString(arr []string, n int) []string {
	end := MinInt(n, len(arr))
	for i := 0; i < end; i++ {
		r := rand.Intn(len(arr))
		arr[i], arr[r] = arr[r], arr[i]
	}
	return arr[:end]
}

func MinInt(x, y int) int {
	if x < y {
		return x
	}

	return y
}

func AbsInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func FileMD5(file string) (string, error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	md5sum := md5.Sum(buf)
	md5sumstr := fmt.Sprintf("%x", md5sum)

	return md5sumstr, nil
}

func StringMD5(str string) string {
	md5sum := md5.Sum([]byte(str))
	md5sumstr := fmt.Sprintf("%x", md5sum)

	return md5sumstr
}

func ChIsLower(ch byte) bool {
	return 'a' <= ch && ch <= 'z'
}

func UpperFirstCharOfWord(word string) string {
	if len(word) == 0 {
		return word
	}

	if !ChIsLower(word[0]) {
		return word
	}

	buf := []byte(word)
	buf[0] -= 'a' - 'A'

	return string(buf)
}

// func RandomNum64(minNum int64, maxNum int64) int64 {
// 	return rand.Int63n(maxNum-minNum+1) + minNum
// }

func init() {
	rand.Seed(time.Now().Unix())
}

func RandomFetchU32(arr []uint32, n int) []uint32 {
	for i := 0; i < n; i++ {
		r := rand.Intn(len(arr))
		arr[i], arr[r] = arr[r], arr[i]
	}

	return arr[:n]
}

func RandomFetchInt(arr []int, n int) []int {
	for i := 0; i < n; i++ {
		r := rand.Intn(len(arr))
		arr[i], arr[r] = arr[r], arr[i]
	}

	return arr[:n]
}

func MaxInt(x, y int) int {
	if x < y {
		return y
	}

	return x
}

// 100 5  10 15
func SplitIntWithDLimit1(sum, n, minGap, maxGap int) (ret []int) {
	Assert(n*minGap <= sum && sum <= n*maxGap)
	for ; n != 1; n-- {
		maxGap2 := sum - (n-1)*minGap
		minGap2 := sum - (n-1)*maxGap
		// logger.Info("minGap2, maxGap2", minGap2, maxGap2)

		min := MaxInt(minGap, minGap2)
		max := MinInt(maxGap, maxGap2)
		Assert(min <= max)
		rd := RandomNum(min, max, nil)
		sum -= rd

		ret = append(ret, rd)
	}

	ret = append(ret, sum)
	return
}

func avgSplit(sum, N int) []int {
	shang := sum / N
	yu := sum % N

	ret := make([]int, N)
	for i := 0; i < N; i++ {
		j := shang
		if i < yu {
			j++
		}

		ret[i] = j
	}

	return ret
}

func rangeRand(x, y, min, max int) int {
	d1 := x - min
	d2 := max - y

	d := MinInt(d1, d2)

	rd := RandomNum(0, d, nil)

	return rd
}

// 10000,  80 , 50 , 300
func CanSplitIntWithDLimit(sum, n, min, max int) bool {
	ok := n > 0 && (min*n <= sum && sum <= max*n)
	if !ok {
		logger.Error("SplitIntWithDLimit 参数冲突 sum, n, min, max", sum, n, min, max)
	}

	return ok
}

// 10000,  80 , 50 , 300
func SplitIntWithDLimit(sum, n, min, max int) (arr []int, ok bool) {
	// if n == 0 || !(min*n <= sum && sum <= max*n) {
	if !CanSplitIntWithDLimit(sum, n, min, max) {
		ok = false
		logger.Fatal("SplitIntWithDLimit 参数冲突 sum, n, min, max", sum, n, min, max)
		return
	}

	arr = avgSplit(sum, n)

	for iii := 0; iii < 2; iii++ {
		for i := 1; i < len(arr); i += 2 {
			j := i - 1

			d := rangeRand(arr[j], arr[i], min, max)
			arr[j] -= d
			arr[i] += d
		}

		ShuffleIntArr(arr)
	}
	ok = true
	return
}

var DEBUG = false

func CanHoleArrayWithInterval(sum, n, minGap, maxGap int) bool {
	return minGap >= 0 && CanSplitIntWithDLimit(sum, n, minGap+1, maxGap+1)
}

func HoleArrayWithInterval(sum, n, minGap, maxGap int) (ret []bool) {
	ret = make([]bool, sum)

	idxArr, ok := SplitIntWithDLimit(sum, n, minGap+1, maxGap+1)
	if !ok {
		return
	}

	if DEBUG {
		logger.Info("idxArr:", idxArr)
	}

	placeholder := idxArr[0] / 3
	idxArr[0] -= placeholder

	if DEBUG {
		logger.Info("idxArr:", idxArr)
	}

	var m int
	for _, v := range idxArr {
		m += v
		ret[m-1] = true
	}

	return
}

type PrioHelper struct {
	Arr []int
	Sum int
}

func (m *PrioHelper) Rand() int {
	n := rand.Intn(m.Sum)

	idx := UpperBoundIdx(m.Arr, n)
	return idx
}

func NewPrioHelper(prioArr []int) *PrioHelper {
	var arr = make([]int, len(prioArr))

	var sum int
	for i, v := range prioArr {
		sum += v
		arr[i] = sum
	}

	if sum == 0 {
		log.Panicln("RandomFromPrioArr: sum is zero")
		return nil
	}

	return &PrioHelper{
		Arr: arr,
		Sum: sum,
	}
}

func RandomFromBalancePrioArr(length int) int {
	arr := []int{}
	for i := 0; i < length; i++ {
		arr = append(arr, 1)
	}
	return RandomFromPrioArr(arr)
}

func RandomFromPrioArr(prioArr []int) int {
	m := NewPrioHelper(prioArr)
	return m.Rand()
}

func Lshift(arr []int) {
	if len(arr) <= 1 {
		return
	}

	first := arr[0]
	for i := 1; i < len(arr); i++ {
		arr[i-1] = arr[i]
	}

	arr[len(arr)-1] = first
}

const stat_YMD_layout = "2006-01-02"

func Time2YMD(t time.Time) string {
	return t.Format(stat_YMD_layout)
}

func YMD2Time(str string) (t time.Time, err error) {
	// return t.Format(stat_YMD_layout)
	t, err = time.Parse(stat_YMD_layout, str)
	return
}

func ToBool(value interface{}) bool {
	boolValue, ok := value.(bool)
	if ok {
		return boolValue
	}
	if value == "true" {
		return true
	}
	return false
}

func CalculateDistance3D(x1, y1, z1, x2, y2, z2 float64) float64 {
	return math.Sqrt(math.Pow(x2-x1, 2)+math.Pow(y2-y1, 2)) + math.Pow(z2-z1, 2)
}

func ToInt(i interface{}) int {
	switch v := i.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case string:
		intV, err := strconv.Atoi(v)
		if err != nil {
			panic("convert is error:" + err.Error())
		} else {
			return intV
		}
	case int64:
		return int(v)
	case float64:
		return int(v)
	case float32:
		return int(v)
	}
	return 0
}

func StrToBase64(i string) string {
	return base64.RawStdEncoding.EncodeToString([]byte(i))
}

func IntToBase64(i int) string {
	str := strconv.Itoa(i)
	return base64.RawStdEncoding.EncodeToString([]byte(str))
}

func Base64ToInt(i string) (n int, e error) {
	var b []byte
	b, e = base64.RawStdEncoding.DecodeString(i)
	if e != nil {
		return
	}

	n, e = strconv.Atoi(string(b))
	return
}

func Base64ToStr(i string) (string, error) {
	b, e := base64.RawStdEncoding.DecodeString(i)
	return string(b), e
}
func ToString(s interface{}) string {
	switch v := s.(type) {
	case int:
		return strconv.Itoa(v)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case float64:
		return strconv.FormatFloat(v, 'f', 2, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', 2, 32)
	case int64:
		if intValue, ok := s.(int64); ok {
			return strconv.FormatInt(intValue, 10)
		}
	case string:
		return v
	}
	return ""
}

func ArrayMerge(dst, src []int) {
	Assert(len(dst) == len(src))

	for i := 0; i < len(dst); i++ {
		dst[i] += src[i]
	}
}

func ArrayMultiple(dst []int, mul int) {
	for i := 0; i < len(dst); i++ {
		dst[i] *= mul
	}
}

func WriteToJsonFile(filename string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "\t")
	return enc.Encode(data)
}

func ParseJsonFile(filename string, data interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewDecoder(file)
	return enc.Decode(data)
}

func GetBoDong(arr []float64) float64 {
	var sum float64
	for _, v := range arr {
		sum += v
	}

	var size = float64(len(arr))

	avg := sum / size

	var fangcha float64
	for _, v := range arr {
		// fangcha += math.Abs(v - avg)
		fangcha += (v - avg) * (v - avg)
	}

	return fangcha / size
}
func Array2Map(arr interface{}, mapkey func(item interface{}) string) map[string]interface{} {
	m := make(map[string]interface{}, 0)
	v := reflect.ValueOf(arr)
	if v.Kind() != reflect.Slice {
		panic("to slice arr not slice")
	}
	l := v.Len()
	for i := 0; i < l; i++ {
		tmpArr := v.Index(i).Interface()
		m[mapkey(tmpArr)] = tmpArr
	}
	return m
}

func MidnightTimeBy(now time.Time) time.Time {
	banye := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return banye
}

func CheckPortAvailability(port int) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		// 端口已被占用
		return false
	}
	defer listener.Close()

	// 端口可用
	return true
}

func GetOutBoundIP() (ip string, err error) {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		logger.Info(err)
		return
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	logger.Info(localAddr.String())
	ip = strings.Split(localAddr.String(), ":")[0]
	return
}

func GetAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	address := listener.Addr().(*net.TCPAddr)
	port := address.Port
	return port, nil
}

// 获取URL的文件名
func GetFileNameFromURL(fileURL string) string {
	// 解析 URL
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return ""
	}

	// 获取 URL 中的路径部分并提取文件名
	fileName := path.Base(parsedURL.Path)

	// 如果 URL 中没有文件名，返回空字符串
	if fileName == "" {
		return "unknown_filename"
	}

	return fileName
}

// 获取来自文件路径的文件名
func GetApplicationName(filePath string) string {
	// 提取文件名（包括扩展名）
	fileName := filepath.Base(filePath)
	// 移除扩展名
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

func MidnightTime() time.Time {
	now := time.Now()
	return MidnightTimeBy(now)
}

func LastSundayMidnightTimeBy(t time.Time) time.Time {
	midnightTime := MidnightTimeBy(t)
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return midnightTime.AddDate(0, 0, 1-weekday)
}

func LastSundayMidnightTime() time.Time {
	return LastSundayMidnightTimeBy(time.Now())
}

func WebSocketPerClose(ws *websocket.Conn, closeCode int, message string) {
	closeMessage := websocket.FormatCloseMessage(closeCode, message)
	deadline := time.Now().Add(time.Second)
	ws.WriteControl(websocket.CloseMessage, closeMessage, deadline)
}

// 年中的第几天
func GetDaysInYearByThisYear() int {
	now := time.Now()
	total := 0
	arr := []int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	y, month, d := now.Date()
	m := int(month)
	for i := 0; i < m-1; i++ {
		total = total + arr[i]
	}
	if (y%400 == 0 || (y%4 == 0 && y%100 != 0)) && m > 2 {
		total = total + d + 1

	} else {
		total = total + d
	}
	return total
}

// 生成单号
// 06123xxxxx
// sum 最少10位,sum 表示全部单号位数
func MakeYearDaysRand(sum int) string {
	//年
	strs := time.Now().Format("06")
	//一年中的第几天
	days := strconv.Itoa(GetDaysInYearByThisYear())
	count := len(days)
	if count < 3 {
		//重复字符0
		days = strings.Repeat("0", 3-count) + days
	}
	//组合
	strs += days
	//剩余随机数
	sum = sum - 5
	if sum < 1 {
		sum = 5
	}
	//0~9999999的随机数
	//ran := GetRand()
	//ran := RandomNum(0, 9999999, nil)
	pow := math.Pow(10, float64(sum)) - 1
	//logger.Info("sum=>", sum)
	//logger.Info("pow=>", pow)
	result := strconv.Itoa(rand.Intn(int(pow)))
	count = len(result)
	//logger.Info("result=>", result)
	if count < sum {
		//重复字符0
		result = strings.Repeat("0", sum-count) + result
	}
	//组合
	strs += result
	return strs
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true // 文件存在
	}
	if os.IsNotExist(err) {
		return false // 文件不存在
	}
	return false // 其他错误
}

func GetMacAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		// 排除虚拟、回环和无效的接口
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		// 获取接口的硬件地址
		mac := iface.HardwareAddr.String()
		if mac != "" {
			return mac, nil
		}
	}

	return "", fmt.Errorf("未找到有效的物理网卡")
}

func DecomposeInteger(n int) (int, int) {
	sqrtN := int(math.Sqrt(float64(n)))

	for i := sqrtN; i > 0; i-- {
		if n%i == 0 {
			num1 := i
			num2 := n / i
			if num1*num2 == n {
				return num1, num2
			}
		}
	}

	return 0, 0
}

func unzipFile(zipFile, destDir string) error {
	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		filePath := filepath.Join(destDir, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		outFile, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, fileReader)
		if err != nil {
			return err
		}
	}

	return nil
}

func downloadFile(url, filePath string) error {
	// 创建一个 HTTP 请求
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 创建文件并将下载内容写入
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

func UpdateApplicationPackage(url string, programPath string) error {
	zipFile := "program.zip"
	destDir := programPath
	os.RemoveAll(destDir)
	os.Remove(zipFile)
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		err := os.Mkdir(destDir, 0755) // 0755 是文件夹权限
		if err != nil {
			return err
		}
	} else {
		return err
	}
	if downloadFile(url, zipFile) != nil {
		return fmt.Errorf("下载失败")
	}
	err := unzipFile(zipFile, destDir)
	if err != nil {
		return fmt.Errorf("解压失败")
	}

	os.Remove(zipFile)
	return nil
}
