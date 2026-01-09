module View.UserDate exposing (view)

import Element exposing (Attribute, Element, px)
import Element.Border as Border
import Element.Font as Font
import Feed exposing (Feed, Post)
import Html
import Html.Attributes as Attr


postText : String -> Element msg
postText text =
    Element.html <|
        Html.span
            [ Attr.style "white-space" "pre-wrap"
            , Attr.style "margin" "10px"
            ]
            [ Html.text text ]


last : List a -> Maybe a
last xs =
    case xs of
        [] ->
            Nothing

        y :: [] ->
            Just y

        _ :: ys ->
            last ys


postUrl : Post -> String
postUrl post =
    let
        p =
            Maybe.withDefault "" <| last <| String.split "/" post.uri
    in
    String.concat
        [ "https://bsky.app/profile/"
        , post.author.did
        , "/post/"
        , p
        ]


toHMS : String -> ( Int, Int, Int )
toHMS ts =
    let
        toIntTuple x =
            case x of
                sh :: sm :: ss :: _ ->
                    let
                        a =
                            ( String.toInt sh
                            , String.toInt sm
                            , Maybe.andThen String.toInt <| List.head <| String.split "." ss
                            )
                    in
                    case a of
                        ( Just h, Just m, Just s ) ->
                            ( h, m, s )

                        _ ->
                            ( 0, 0, 0 )

                _ ->
                    ( 0, 0, 0 )
    in
    Maybe.withDefault ( 0, 0, 0 ) <|
        Maybe.map (toIntTuple << String.split ":" << String.dropRight 1) <|
            last <|
                String.split "T" ts


fromInt02 : Int -> String
fromInt02 i =
    let
        p =
            if i < 10 then
                "0"

            else
                ""
    in
    p ++ String.fromInt i


formatHMS : ( Int, Int ) -> ( Int, Int, Int ) -> String
formatHMS ( zh, zm ) hm =
    case hm of
        ( 0, 0, 0 ) ->
            "(unknown time)"

        ( h, m, s ) ->
            String.concat
                [ fromInt02 (modBy 24 (h + 24 + zh))
                , ":"
                , fromInt02 (modBy 60 (m + 60 + zm))
                , ":"
                , fromInt02 s
                ]


viewFeed : Feed -> Element msg
viewFeed feed =
    Element.row
        [ Element.spacing 10
        , Border.widthEach
            { bottom = 0
            , left = 0
            , right = 0
            , top = 1
            }
        , Element.paddingXY 0 10
        ]
        [ Element.image
            [ Element.width (px 48)
            , Element.height (px 48)
            , Element.alignTop
            ]
            { src = feed.post.author.avatar
            , description = ""
            }
        , Element.column
            [ Element.width (px 600)
            ]
            [ Element.row [ Element.spacing 5 ]
                [ Element.text feed.post.author.displayName
                , Element.text <| "@" ++ feed.post.author.handle
                ]
            , postText feed.post.record.text
            , Element.link [ Font.size 15 ]
                { url = postUrl feed.post
                , label =
                    Element.text <|
                        formatHMS ( 9, 0 ) <|
                            toHMS <|
                                feed.post.record.createdAt
                }
            ]
        ]


view : String -> String -> String -> String -> List Feed -> Element msg
view user year month day feeds =
    Element.column
        [ Element.spacing 10
        , Element.paddingXY 30 0
        ]
        [ Element.row []
            [ Element.text "Date:"
            , Element.row []
                [ Element.text year
                , Element.text "/"
                , Element.text month
                , Element.text "/"
                , Element.text day
                ]
            ]
        , Element.column [] <| List.map viewFeed feeds
        ]
