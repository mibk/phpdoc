package phpdoc_test

import (
	"strings"
	"testing"

	"mibk.dev/phpdoc"
)

var printTests = []struct {
	name string
	test string
}{
	{"basic", `
/**
	@param array &  $x
	@param string&  ... $bar
	@return ? float
*/
----
/**
 * @param  array  &$x
 * @param  string &...$bar
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
@param  bool... $opts
 *@return float    Always positive
*/
----
   /**
    * @author Name <not known>
    * @param  DateTime|string|null $bar Must be   from this century
    * @param  mixed                $foo
    * @param  bool                 ...$opts
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
@implements   Comparable <self ,>
@extends \ DateTime
@uses \ HandyOne
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
 * @implements Comparable<self>
 * @extends    \DateTime
 * @uses       \HandyOne
 */
`},
	{"properties", `
	/**
@property  \ Foo $a Foo
@property-read    array<int,string>    $b   bar
@property-write int [] $c
@property array    {0 :int  ,foo?:\ Foo ,}$d
*/
----
	/**
	 * @property       \Foo                      $a Foo
	 * @property-read  array<int, string>        $b bar
	 * @property-write int[]                     $c
	 * @property       array{0: int, foo?: \Foo} $d
	 */
`},
	{"array shapes", `
/**
@var  array{0  ? :int, one :string,}
@var array{'foo': string ,'bar\'' ?: string}
*/
----
/**
 * @var array{0?: int, one: string}
 * @var array{'foo': string, 'bar\''?: string}
 */
`},
	{"object shapes", `
/**
@var  object   {name :string ,role  ?: int ,}
@return        object
*/
----
/**
 * @var    object{name: string, role?: int}
 * @return object
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
	{"callable", `
/**
@param  callable$m
@param  callable(): string $k
@param callable  ( int  $a ,string& ...$b, ): string |int$x
@param  callable (bool ,string ) $n
@param  callable (string$s='false' ,bool $b  =true, ) $nn
@param callable(  ) :$this$m
@param int... $y
@return callable  ( int $a , int $b=3 ) :void
*/
----
/**
 * @param  callable                                      $m
 * @param  callable(): string                            $k
 * @param  callable(int $a, string &...$b): string|int   $x
 * @param  callable(bool, string)                        $n
 * @param  callable(string $s = 'false', bool $b = true) $nn
 * @param  callable(): $this                             $m
 * @param  int                                           ...$y
 * @return callable(int $a, int $b = 3): void
 */
`},
	{"method tag", `
/**
@method  ? \ DateTime getDate( int| string$c ,)   the date of x
@method translate (mixed& ... $args)does that for y
@method  static  void  clean ( )
*/
----
/**
 * @method ?\DateTime getDate(int|string $c) the date of x
 * @method translate(mixed &...$args)        does that for y
 * @method static void clean()
 */
`},
	{"const exprs", `
/**
@param   self :: ALL_*$a
@param  self::ANY_*  []           $xx
@param value-of < That   :: ConstVal >$b
@return BAR  ::    *
*/
----
/**
 * @param  self::ALL_*              $a
 * @param  self::ANY_*[]            $xx
 * @param  value-of<That::ConstVal> $b
 * @return BAR::*
 */
`},
	{"types", `
/**
@phpstan-type   Foo  array{ 'bar' :string }  For sure
@param  Foo     []   $foos
*/
----
/**
 * @phpstan-type Foo   array{'bar': string} For sure
 * @param        Foo[] $foos
 */
`},
	{"string lit", `
/**
@param  'foo' |  7 | 'bar'     []   $xyz
*/
----
/**
 * @param 'foo'|7|'bar'[] $xyz
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
