module Csv.Parser exposing (file)

import Parser as P exposing ((|.), (|=), Parser)


comma : Parser ()
comma =
    P.symbol ","


dquote : Parser ()
dquote =
    P.symbol "\""


cr : Parser ()
cr =
    P.symbol "\u{000D}"


lf : Parser ()
lf =
    P.symbol "\n"


crlf : Parser ()
crlf =
    cr |> P.andThen (always lf)


newline : Parser ()
newline =
    P.oneOf [ crlf, lf ]


nonEscaped : Parser String
nonEscaped =
    P.getChompedString <| P.chompWhile <| \c -> not <| List.member c (String.toList ",\u{000D}\n")


escaped : Parser String
escaped =
    P.succeed identity
        |. dquote
        |= P.getChompedString (P.chompWhile ((/=) '"'))
        |. dquote


field : Parser String
field =
    P.oneOf [ escaped, nonEscaped ]


many : Parser a -> Parser (List a)
many p =
    P.loop []
        (\accm ->
            P.oneOf
                [ P.map (\a -> P.Loop (a :: accm)) p
                , P.map (always <| P.Done (List.reverse accm)) <| P.succeed ()
                ]
        )


sepBy1 : Parser a -> Parser sep -> Parser (List a)
sepBy1 p sep =
    P.andThen
        (\x ->
            P.andThen
                (\xs -> P.succeed (x :: xs))
                (many (sep |> P.andThen (always p)))
        )
        p


record : Parser (List String)
record =
    sepBy1 field comma


optional : Parser a -> Parser ()
optional p =
    P.oneOf
        [ P.map (always ()) p
        , P.succeed ()
        ]


file : Parser (List (List String))
file =
    sepBy1 record newline
        |> P.andThen
            (\ls ->
                optional newline |> P.andThen (always <| P.succeed ls)
            )
        -- drop header
        |> P.map (List.drop 1)
