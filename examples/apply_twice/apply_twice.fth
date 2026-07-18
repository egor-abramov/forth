import io

var callback

: apply_twice (2 -> 1)
    callback !
    callback @ execute
    callback @ execute
;

: double (1 -> 1)
    2 *
;

: square (1 -> 1)
    dup *
;

string "DoubleTest" str_msg1
string "SquareTest" str_msg2

str_msg1 print_str cr
5 ' double apply_twice
. cr

str_msg2 print_str cr
3 ' square apply_twice
.