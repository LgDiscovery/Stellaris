package node

import "fmt"

func reverseString(s []byte) {
	left, right := 0, len(s)-1
	for left < right {
		s[left], s[right] = s[right], s[left]
		left++
		right--
	}
}

func isAnagram(s string, t string) bool {
	if len(s) != len(t) {
		return false
	}
	cnt := [26]int{}
	for _, c := range s {
		cnt[c-'a']++
	}
	for _, c := range t {
		cnt[c-'a']--
		if cnt[c-'a'] < 0 {
			return false
		}
	}
	return true
}

func longestPalindrome(s string) string {
	res := ""
	// 中心扩展法
	var expand func(l, r int)
	expand = func(l, r int) {
		for l >= 0 && r < len(s) && s[l] == s[r] {
			if r-l+1 > len(res) {
				res = s[l : r+1]
			}
			l--
			r++
		}
	}
	for i := range s {
		expand(i, i)   // 奇数长度
		expand(i, i+1) // 偶数长度
	}
	return res
}

func strStr(haystack string, needle string) int {
	n, m := len(haystack), len(needle)
	for i := 0; i <= n-m; i++ {
		j := 0
		for j < m && haystack[i+j] == needle[j] {
			j++
		}
		if j == m {
			return i
		}
	}
	return -1
}

func climbStairs(n int) int {
	if n <= 2 {
		return n
	}
	a, b := 1, 2
	for i := 3; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

func maxSubArray(nums []int) int {
	dp := nums[0]
	res := dp
	for i := 1; i < len(nums); i++ {
		dp = max(nums[i], dp+nums[i])
		res = max(res, dp)
	}
	return res
}

func rob(nums []int) int {
	n := len(nums)
	if n == 0 {
		return 0
	}
	if n == 1 {
		return nums[0]
	}
	dp0, dp1 := nums[0], max(nums[0], nums[1])
	for i := 2; i < n; i++ {
		dp := max(dp1, dp0+nums[i])
		dp0, dp1 = dp1, dp
	}
	return dp1
}

func subsets(nums []int) [][]int {
	var res [][]int
	var path []int
	var backtrack func(start int)
	backtrack = func(start int) {
		// 加入结果
		temp := make([]int, len(path))
		copy(temp, path)
		res = append(res, temp)
		// 遍历选择
		for i := start; i < len(nums); i++ {
			path = append(path, nums[i])
			backtrack(i + 1)
			path = path[:len(path)-1] // 回溯
		}
	}
	backtrack(0)
	return res
}

func permute(nums []int) [][]int {
	var res [][]int
	var path []int
	used := make([]bool, len(nums))
	var backtrack func()
	backtrack = func() {
		if len(path) == len(nums) {
			temp := make([]int, len(path))
			copy(temp, path)
			res = append(res, temp)
			return
		}
		for i := 0; i < len(nums); i++ {
			if used[i] {
				continue
			}
			used[i] = true
			path = append(path, nums[i])
			backtrack()
			path = path[:len(path)-1]
			used[i] = false
		}
	}
	backtrack()
	return res
}

// 邻接表表示图
func dfs(graph [][]int, start int, visited []bool) {
	visited[start] = true
	// 访问节点
	fmt.Println(start)
	// 遍历邻接节点
	for _, v := range graph[start] {
		if !visited[v] {
			dfs(graph, v, visited)
		}
	}
}

func bfs(graph [][]int, start int) {
	n := len(graph)
	visited := make([]bool, n)
	queue := []int{start}
	visited[start] = true
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		fmt.Println(node)
		for _, v := range graph[node] {
			if !visited[v] {
				visited[v] = true
				queue = append(queue, v)
			}
		}
	}
}

func topologicalSort(numCourses int, prerequisites [][]int) []int {
	// 入度+邻接表
	inDegree := make([]int, numCourses)
	graph := make([][]int, numCourses)
	for _, p := range prerequisites {
		graph[p[1]] = append(graph[p[1]], p[0])
		inDegree[p[0]]++
	}
	// 队列存入度为0的节点
	queue := []int{}
	for i := 0; i < numCourses; i++ {
		if inDegree[i] == 0 {
			queue = append(queue, i)
		}
	}
	res := []int{}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		res = append(res, node)
		// 邻接节点入度-1
		for _, v := range graph[node] {
			inDegree[v]--
			if inDegree[v] == 0 {
				queue = append(queue, v)
			}
		}
	}
	// 有环返回空
	if len(res) != numCourses {
		return []int{}
	}
	return res
}
