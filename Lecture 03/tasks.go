package homework

// PairInput соответствует объекту nums/target из условия контеста.
type PairInput struct {
	Nums   []int
	Target int
}

// longestPrefix возвращает самый длинный общий префикс строк.
func longestPrefix(words []string) string {
	if len(words) == 0 {
		return ""
	}

	prefix := []rune(words[0])

	for _, word := range words[1:] {
		if len(prefix) == 0 {
			return ""
		}

		runeWord := []rune(word)

		i := 0

		for i < len(runeWord) && i < len(prefix) && prefix[i] == runeWord[i] {
			i++
		}

		prefix = prefix[:i]
	}

	return string(prefix)
}

// pairIndices возвращает индексы двух разных элементов по возрастанию.
func pairIndices(input PairInput) [2]int {
	prevElems := make(map[int]int, len(input.Nums))
	for i, v := range input.Nums {
		if prevIndex, ok := prevElems[input.Target-v]; ok {
			return [2]int{prevIndex, i}
		}

		prevElems[v] = i
	}
	return [2]int{}
}

// hasDuplicate сообщает, встречается ли какое-либо значение повторно.
func hasDuplicate(nums []int) bool {
	seen := make(map[int]bool, len(nums))

	for _, v := range nums {
		if seen[v] {
			return true
		}
		seen[v] = true
	}

	return false
}

// compactSorted создаёт независимый слайс уникальных значений.
func compactSorted(nums []int) []int {
	if len(nums) == 0 {
		return nil
	}

	res := make([]int, 0, len(nums))
	res = append(res, nums[0])

	for j := 1; j < len(nums); j++ {
		if nums[j-1] != nums[j] {
			res = append(res, nums[j])
		}
	}

	return res
}

// balancedBrackets проверяет пары и порядок закрытия скобок.
func balancedBrackets(text string) bool {
	bracketsStack := make([]rune, 0, len(text))

	for _, b := range text {
		switch b {
		case '(':
			bracketsStack = append(bracketsStack, ')')
		case '[':
			bracketsStack = append(bracketsStack, ']')
		case '{':
			bracketsStack = append(bracketsStack, '}')
		case ')', ']', '}':
			if len(bracketsStack) == 0 || bracketsStack[len(bracketsStack)-1] != b {
				return false
			}
			bracketsStack = bracketsStack[:len(bracketsStack)-1]
		}

	}
	return len(bracketsStack) == 0
}
