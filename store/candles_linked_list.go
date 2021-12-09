package store

import (
	"fmt"

	"github.com/mikekonan/freqtradeProxy/model"
)

type candlesLinkedList struct {
	first *element
	last  *element
	len   int
}

type element struct {
	value *model.Candle
	prev  *element
	next  *element
}

func newCandlesLinkedList(values ...*model.Candle) *candlesLinkedList {
	list := &candlesLinkedList{}
	if len(values) > 0 {
		list.add(values...)
	}
	return list
}

func (list *candlesLinkedList) add(values ...*model.Candle) {
	for _, value := range values {
		newElement := &element{value: value, prev: list.last}
		if list.len == 0 {
			list.first = newElement
			list.last = newElement
		} else {
			list.last.next = newElement
			list.last = newElement
		}
		list.len++
	}
}

func (list *candlesLinkedList) append(values ...*model.Candle) {
	list.add(values...)
}

func (list *candlesLinkedList) prepend(values ...*model.Candle) {

	for v := len(values) - 1; v >= 0; v-- {
		newElement := &element{value: values[v], next: list.first}
		if list.len == 0 {
			list.first = newElement
			list.last = newElement
		} else {
			list.first.prev = newElement
			list.first = newElement
		}
		list.len++
	}
}

func (list *candlesLinkedList) get(index int) (*model.Candle, bool) {
	if !list.withinRange(index) {
		return nil, false
	}

	if list.len-index < index {
		element := list.last
		for e := list.len - 1; e != index; e, element = e-1, element.prev {
		}
		return element.value, true
	}

	element := list.first
	for e := 0; e != index; e, element = e+1, element.next {
	}

	return element.value, true
}

func (list *candlesLinkedList) remove(index int) {
	if !list.withinRange(index) {
		return
	}

	if list.len == 1 {
		list.clear()
		return
	}

	var element *element

	if list.len-index < index {
		element = list.last
		for e := list.len - 1; e != index; e, element = e-1, element.prev {
		}
	} else {
		element = list.first
		for e := 0; e != index; e, element = e+1, element.next {
		}
	}

	if element == list.first {
		list.first = element.next
	}
	if element == list.last {
		list.last = element.prev
	}
	if element.prev != nil {
		element.prev.next = element.next
	}
	if element.next != nil {
		element.next.prev = element.prev
	}

	element = nil

	list.len--
}

func (list *candlesLinkedList) contains(values ...*model.Candle) bool {
	if len(values) == 0 {
		return true
	}
	if list.len == 0 {
		return false
	}
	for _, value := range values {
		found := false
		for element := list.first; element != nil; element = element.next {
			if element.value == value {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (list *candlesLinkedList) values() []*model.Candle {
	values := make([]*model.Candle, list.len, list.len)
	for e, element := 0, list.first; element != nil; e, element = e+1, element.next {
		values[e] = element.value
	}
	return values
}

func (list *candlesLinkedList) invertedValues() []*model.Candle {
	values := make([]*model.Candle, list.len, list.len)
	for e, element := 0, list.last; element != nil; e, element = e+1, element.prev {
		values[e] = element.value
	}
	return values
}

func (list *candlesLinkedList) indexOf(value *model.Candle) int {
	if list.len == 0 {
		return -1
	}
	for index, element := range list.values() {
		if element == value {
			return index
		}
	}
	return -1
}

func (list *candlesLinkedList) empty() bool {
	return list.len == 0
}

func (list *candlesLinkedList) size() int {
	return list.len
}

func (list *candlesLinkedList) clear() {
	list.len = 0
	list.first = nil
	list.last = nil
}

func (list *candlesLinkedList) swap(i, j int) {
	if list.withinRange(i) && list.withinRange(j) && i != j {
		var element1, element2 *element
		for e, currentElement := 0, list.first; element1 == nil || element2 == nil; e, currentElement = e+1, currentElement.next {
			switch e {
			case i:
				element1 = currentElement
			case j:
				element2 = currentElement
			}
		}
		element1.value, element2.value = element2.value, element1.value
	}
}

func (list *candlesLinkedList) insert(index int, values ...*model.Candle) {

	if !list.withinRange(index) {

		if index == list.len {
			list.add(values...)
		}
		return
	}

	list.len += len(values)

	var beforeElement *element
	var foundElement *element

	if list.len-index < index {
		foundElement = list.last
		for e := list.len - 1; e != index; e, foundElement = e-1, foundElement.prev {
			beforeElement = foundElement.prev
		}
	} else {
		foundElement = list.first
		for e := 0; e != index; e, foundElement = e+1, foundElement.next {
			beforeElement = foundElement
		}
	}

	if foundElement == list.first {
		oldNextElement := list.first
		for i, value := range values {
			newElement := &element{value: value}
			if i == 0 {
				list.first = newElement
			} else {
				newElement.prev = beforeElement
				beforeElement.next = newElement
			}
			beforeElement = newElement
		}
		oldNextElement.prev = beforeElement
		beforeElement.next = oldNextElement
	} else {
		oldNextElement := beforeElement.next
		for _, value := range values {
			newElement := &element{value: value}
			newElement.prev = beforeElement
			beforeElement.next = newElement
			beforeElement = newElement
		}
		oldNextElement.prev = beforeElement
		beforeElement.next = oldNextElement
	}
}

func (list *candlesLinkedList) set(index int, value *model.Candle) {

	if !list.withinRange(index) {

		if index == list.len {
			list.add(value)
		}
		return
	}

	var foundElement *element

	if list.len-index < index {
		foundElement = list.last
		for e := list.len - 1; e != index; {
			fmt.Println("set last", index, value, foundElement, foundElement.prev)
			e, foundElement = e-1, foundElement.prev
		}
	} else {
		foundElement = list.first
		for e := 0; e != index; {
			e, foundElement = e+1, foundElement.next
		}
	}

	foundElement.value = value
}

func (list *candlesLinkedList) withinRange(index int) bool {
	return index >= 0 && index < list.len
}

func (list *candlesLinkedList) selectInRangeReversedFn(fromSelectorFn func(*model.Candle) bool, toSelectorFn func(*model.Candle) bool) []*model.Candle {
	values := make([]*model.Candle, 0)

	started := false

	for element := list.last; element != nil; element = element.prev {
		if started || toSelectorFn(element.value) {
			started = true
		} else {
			continue
		}

		values = append(values, element.value)

		if started && fromSelectorFn(element.value) {
			break
		}
	}

	return values
}

func (list *candlesLinkedList) selectFn(selectorFn func(*model.Candle) bool) []*model.Candle {
	values := make([]*model.Candle, 0)

	for element := list.first; element != nil; element = element.next {
		if selectorFn(element.value) {
			values = append(values, element.value)
		}
	}

	return values
}

func (list *candlesLinkedList) selectBreakFn(selectorFn func(*model.Candle) bool, breakFn func(*model.Candle, []*model.Candle) bool) []*model.Candle {
	values := make([]*model.Candle, 0)

	for element := list.first; element != nil; element = element.next {
		if breakFn(element.value, values) {
			break
		}

		if selectorFn(element.value) {
			values = append(values, element.value)
		}
	}

	return values
}
