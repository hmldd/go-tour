/*
	A Tour of Go Exercise: Equivalent Binary Trees
	Go语言之旅 - 判断两个二叉树是否相等
	https://tour.golang.org/concurrency/8
*/
package main

import "golang.org/x/tour/tree"

// Walk walks the tree t sending all values
// from the tree to the channel ch.
func Walk(t *tree.Tree, ch chan<- int) {
	if t == nil {
		return
	}

	// 使用递归的方式遍历二叉树的左分支
	if t.Left != nil {
		Walk(t.Left, ch)
	}

	// 将值传递给ch
	ch <- t.Value

	// 遍历右分支
	if t.Right != nil {
		Walk(t.Right, ch)
	}
}

// Same determines whether the trees
// t1 and t2 contain the same values.
func Same(t1, t2 *tree.Tree) bool {
	ch1 := make(chan int)
	go func() {
		// 启动goroutine时使用闭包和defer关闭ch1，
		defer close(ch1)
		Walk(t1, ch1)
	}()

	ch2 := make(chan int)
	go func() {
		defer close(ch2)
		Walk(t2, ch2)
	}()

	for {
		// 使用第二个参数ok判断channel是否关闭
		v1, ok1 := <-ch1
		v2, ok2 := <-ch2

		// 如果当前长度不相等（ok1 != ok2）或当前值不相等返回false
		if ok1 != ok2 || v1 != v2 {
			return false
		}

		// 二叉树的长度和值都相等时，如果ch1已经关闭退出for循环
		if !ok1 {
			break
		}
	}

	return true
}

func main() {
	isSame1 := Same(tree.New(1), tree.New(1))
	println(isSame1)

	isSame2 := Same(tree.New(1), tree.New(2))
	println(isSame2)
}
