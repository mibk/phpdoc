package phpdoc_test

import (
	"strings"
	"testing"

	"mibk.io/phpdoc"
)

var printTests = []struct {
	name string
	test string
}{
	{"basic", `
/**
	@param string ... $bar
	@return ? float
*/
----
/**
 * @param  string ...$bar
 * @return ?float
 */
`},
	{"oneline", `
  /**@var \ DateTime $date    */
----
  /** @var \DateTime $date */
`},
	{"single line", `
  /**
@var \ Traversable*/
----
  /**
   * @var \Traversable
   */
`},
	{"more params", `
   /**
	@author   Name <not known>
@param DateTime | string|null $bar Must be   from this century
@param mixed $foo
 *@return float    Always positive
*/
----
   /**
    * @author Name <not known>
    * @param  DateTime|string|null $bar Must be   from this century
    * @param  mixed                $foo
    * @return float                Always positive
    */
`},
	{"tags and text", `
/**
This function does
* * this and
* * that.

 * @author   Jack
It's	deprecated now.

@deprecated Don't use
@return bool
@throws  \  InvalidArgumentException
*/
----
/**
 * This function does
 * * this and
 * * that.
 *
 * @author Jack
 * It's	deprecated now.
 *
 * @deprecated Don't use
 * @return     bool
 * @throws     \InvalidArgumentException
 */
`},
	{"properties", `
	/**
@property  \ Foo $a
@property-read    array<int,string>    $b
@property-write int [] $c
@property array    {0 :int  ,foo?:\ Foo }$d
*/
----
	/**
	 * @property       \Foo                      $a
	 * @property-read  array<int, string>        $b
	 * @property-write int[]                     $c
	 * @property       array{0: int, foo?: \Foo} $d
	 */
`},
	{"template", `
/**
@template    T foo
@template  U of \ Traversable bar
@template   WW as \ Countable */
----
/**
 * @template T                 foo
 * @template U of \Traversable bar
 * @template WW of \Countable
 */
`},
}

func TestPrinting(t *testing.T) {
	for _, tt := range printTests {
		t.Run(tt.name, func(t *testing.T) {
			s := strings.Split(tt.test, "----\n")
			if len(s) != 2 {
				t.Fatal("invalid test format")
			}

			input, want := s[0], s[1]
			printerTestCase(t, input, want)
		})
	}
}

func printerTestCase(t *testing.T, input, want string) {
	doc, err := phpdoc.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	got := new(strings.Builder)
	if err := phpdoc.Fprint(got, doc); err != nil {
		t.Fatalf("printing: unexpected err: %v", err)
	}
	if got.String() != want {
		t.Errorf("\n got: %s\nwant: %s", got, want)
	}
}
