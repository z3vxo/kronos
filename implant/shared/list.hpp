#pragma once
#include <windows.h>

template <typename T>
struct Node {
	T data;
	Node<T>* next;
};

template <typename T>
class List {
public:
	List() : head(NULL) {};

	Node<T>* Add(const T& item) {
		Node<T>* node = (Node<T>*)HeapAlloc(GetProcessHeap(), HEAP_ZERO_MEMORY, sizeof(Node<T>));
		if (!node) return NULL;
		node->data = item;
		node->next = head;
		head = node;
		return node;
	}

	void Remove(Node<T>* prev, Node<T>* cur) {
		if (prev) {
			prev->next = cur->next;
		}
		else {
			head = cur->next;
		}
		HeapFree(GetProcessHeap(), 0, cur);
	}


	Node<T>* Head() { return head; }

private:
	Node<T>* head;

};
