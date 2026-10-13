package node

func search(nums []int, target int) int {
	left, right := 0, len(nums)-1
	for left <= right {
		mid := left + (right-left)/2 // 防止溢出
		if nums[mid] == target {
			return mid
		} else if nums[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return -1
}

func searchLeft(nums []int, target int) int {
	left, right := 0, len(nums)-1
	for left <= right {
		mid := left + (right-left)/2
		if nums[mid] == target {
			right = mid - 1 // 继续向左找
		} else if nums[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	// 校验边界
	if left < len(nums) && nums[left] == target {
		return left
	}
	return -1
}

func searchRight(nums []int, target int) int {
	left, right := 0, len(nums)-1
	for left <= right {
		mid := left + (right-left)/2
		if nums[mid] == target {
			left = mid + 1 // 继续向右找
		} else if nums[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	if right >= 0 && nums[right] == target {
		return right
	}
	return -1
}

func searchRotate(nums []int, target int) int {
	left, right := 0, len(nums)-1
	for left <= right {
		mid := left + (right-left)/2
		if nums[mid] == target {
			return mid
		}
		// 左半部分有序
		if nums[left] <= nums[mid] {
			if nums[left] <= target && target < nums[mid] {
				right = mid - 1
			} else {
				left = mid + 1
			}
		} else { // 右半部分有序
			if nums[mid] < target && target <= nums[right] {
				left = mid + 1
			} else {
				right = mid - 1
			}
		}
	}
	return -1
}

func nextGreaterElement(nums []int) []int {
	res := make([]int, len(nums))
	stack := []int{} // 存索引
	for i := len(nums) - 1; i >= 0; i-- {
		// 栈顶小于当前值，弹出
		for len(stack) > 0 && nums[stack[len(stack)-1]] <= nums[i] {
			stack = stack[:len(stack)-1]
		}
		if len(stack) == 0 {
			res[i] = -1
		} else {
			res[i] = nums[stack[len(stack)-1]]
		}
		stack = append(stack, i)
	}
	return res
}

func trap(height []int) int {
	res := 0
	stack := []int{} // 存索引
	for i := range height {
		// 形成凹槽
		for len(stack) > 0 && height[i] > height[stack[len(stack)-1]] {
			top := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if len(stack) == 0 {
				break
			}
			left := stack[len(stack)-1]
			w := i - left - 1
			h := min(height[left], height[i]) - height[top]
			res += w * h
		}
		stack = append(stack, i)
	}
	return res
}

func largestRectangleArea(heights []int) int {
	res := 0
	// 首尾加0，方便处理边界
	newHeights := append([]int{0}, heights...)
	newHeights = append(newHeights, 0)
	stack := []int{}
	for i := range newHeights {
		for len(stack) > 0 && newHeights[i] < newHeights[stack[len(stack)-1]] {
			top := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			h := newHeights[top]
			w := i - stack[len(stack)-1] - 1
			res = max(res, h*w)
		}
		stack = append(stack, i)
	}
	return res
}

func lengthOfLongestSubstring(s string) int {
	m := make(map[byte]int) // 存字符最后出现位置
	left := 0
	res := 0
	for right := range s {
		// 窗口收缩
		if m[s[right]] > 0 {
			left = max(left, m[s[right]])
		}
		m[s[right]] = right + 1
		res = max(res, right-left+1)
	}
	return res
}

func minSubArrayLen(target int, nums []int) int {
	n := len(nums)
	res := n + 1
	sum := 0
	left := 0
	for right := 0; right < n; right++ {
		sum += nums[right]
		// 收缩窗口
		for sum >= target {
			res = min(res, right-left+1)
			sum -= nums[left]
			left++
		}
	}
	if res == n+1 {
		return 0
	}
	return res
}

func findAnagrams(s string, p string) []int {
	var res []int
	need, window := [26]int{}, [26]int{}
	for _, c := range p {
		need[c-'a']++
	}
	left := 0
	for right := range s {
		window[s[right]-'a']++
		// 窗口大小等于p长度
		if right-left+1 == len(p) {
			if need == window {
				res = append(res, left)
			}
			window[s[left]-'a']--
			left++
		}
	}
	return res
}
