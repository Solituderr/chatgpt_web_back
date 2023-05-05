package main

import (
	"fmt"

	"github.com/pandodao/tokenizer-go"
)

func Test() {
	t := tokenizer.MustCalToken(`Many words map to one token, but some don't: indivisible.

Unicode characters like emojis may be split into many tokens containing the underlying bytes: 🤚🏾

Sequences of characters commonly found next to each other may be grouped together: 1234567890`)
	fmt.Println(t) // Output: 64

	// Output: {Bpe:[7085 2456 3975 284 530 11241] Text:[Many  words  map  to  one  token]}
	fmt.Printf("%+v\n", tokenizer.MustEncode("Many words map to one token"))

	// Output: Many words map to one token
	fmt.Println(tokenizer.MustDecode([]int{7085, 2456, 3975, 284, 530, 11241}))
}
