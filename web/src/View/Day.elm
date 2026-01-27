module View.Day exposing
    ( Model
    , Msg(..)
    , init
    , update
    , view
    )

import Element exposing (Attribute, Element, px)
import Element.Border as Border
import Element.Font as Font
import Feed exposing (Embed, External, Facet, Feed, Image, Post, Record, Reply, Video)
import Html exposing (Html)
import Html.Attributes as Attr
import Lib
import String.UTF8 as UTF8
import Task


type alias Model =
    { user : String
    , year : String
    , month : String
    , day : String
    , feeds : List Feed
    }


init : Model
init =
    Model "" "" "" "" []


type Msg
    = UpdateDay String String String String
    | UpdateFeeds (List Feed)
    | FetchFeeds String String String String


update : (Msg -> msg) -> Msg -> Model -> ( Model, Cmd msg )
update toMsg msg model =
    case msg of
        UpdateDay user year month day ->
            if
                Lib.any
                    [ user /= model.user
                    , year /= model.year
                    , month /= model.month
                    , day /= model.day
                    ]
            then
                ( { model
                    | user = user
                    , year = year
                    , month = month
                    , day = day
                  }
                , Lib.perform toMsg <| FetchFeeds user year month day
                )

            else
                ( model, Cmd.none )

        UpdateFeeds feeds ->
            ( { model | feeds = feeds }, Cmd.none )

        _ ->
            ( model, Cmd.none )


type Text
    = String String
    | Tag String
    | Link String String


split : Int -> Int -> List a -> List a
split start end =
    List.drop start >> List.take (end - start)


toS : List Int -> String
toS bs =
    case UTF8.toString bs of
        Ok s ->
            s

        Err err ->
            ""


decorate : Record -> List Text
decorate record =
    let
        bs =
            UTF8.toBytes record.text

        f start facets =
            case facets of
                facet :: fs ->
                    String (toS <| split start facet.index.start bs)
                        :: (let
                                txt =
                                    toS <| split facet.index.start facet.index.end bs
                            in
                            case List.head facet.features of
                                Just (Feed.Tag tag) ->
                                    Tag tag

                                Just (Feed.Link url) ->
                                    Link txt url

                                Nothing ->
                                    String txt
                           )
                        :: f facet.index.end fs

                [] ->
                    case List.drop start bs of
                        [] ->
                            []

                        bs_ ->
                            [ String <| toS bs_ ]
    in
    f 0 record.facets


viewText : List Text -> List (Html msg)
viewText list =
    case list of
        t :: ts ->
            (case t of
                String s ->
                    Html.text s

                Tag tag ->
                    Html.a
                        [ Attr.href <| "https://bsky.app/hashtag/" ++ tag ]
                        [ Html.text <| "#" ++ tag ]

                Link caption url ->
                    Html.a
                        [ Attr.href url ]
                        [ Html.text caption ]
            )
                :: viewText ts

        [] ->
            []


postText : Record -> Element msg
postText record =
    if String.isEmpty record.text then
        Element.none

    else
        Element.html <|
            Html.span
                [ Attr.style "white-space" "pre-wrap"
                , Attr.style "margin" "10px"
                ]
            <|
                viewText <|
                    decorate record


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


imageView : Image -> Element msg
imageView image =
    Element.link []
        { url = image.fullsize
        , label =
            Element.image
                [ Element.width (px 140) ]
                { src = image.thumb
                , description = image.alt
                }
        }


externalView : External -> Element msg
externalView external =
    Element.link []
        { url = external.uri
        , label =
            Element.column
                [ Border.rounded 5
                , Border.width 1
                , Element.width (px 410)
                ]
                [ Element.el [ Element.padding 5 ] <| Element.text external.title
                , Element.image
                    [ Element.width (px 400)
                    , Element.padding 5
                    ]
                    { src = external.thumb
                    , description = external.description
                    }
                ]
        }


videoView : Video -> Element msg
videoView video =
    Element.link []
        { url = video.playlist
        , label =
            Element.image
                [ Element.width (px 140) ]
                { src = video.thumbnail
                , description = "embedded video"
                }
        }


embedView : Embed -> Element msg
embedView embed =
    Element.el [ Element.paddingXY 10 0 ] <|
        case embed of
            Feed.EmbeddedImage images ->
                Element.row [ Element.spacingXY 0 3 ] <|
                    List.map imageView images

            Feed.EmbeddedVideo video ->
                videoView video

            Feed.EmbeddedExternal external ->
                externalView external

            Feed.EmbeddedPost post ->
                Element.el
                    [ Border.rounded 3
                    , Border.width 1
                    , Element.width (px 550)
                    , Font.size 17
                    ]
                <|
                    Element.el [ Element.padding 3 ] <|
                        viewPost True post Nothing


viewPost : Bool -> Post -> Maybe Reply -> Element msg
viewPost small post reply =
    let
        iconSize =
            if small then
                36

            else
                48

        width =
            if small then
                500

            else
                600

        timestampSize =
            if small then
                12

            else
                15
    in
    Element.row
        [ Element.spacing 10
        ]
        [ Element.link [ Element.alignTop ]
            { url = "https://bsky.app/profile/" ++ post.author.handle
            , label =
                Element.image
                    [ Element.width (px iconSize)
                    , Element.height (px iconSize)
                    ]
                    { src = post.author.avatar
                    , description = ""
                    }
            }
        , Element.column
            [ Element.width (px width)
            , Element.spacing 10
            ]
            [ Lib.maybe Element.none
                (\r ->
                    if r.parent.cid == r.root.cid then
                        Element.none

                    else
                        Element.link []
                            { url = postUrl r.root
                            , label =
                                Element.el
                                    [ Element.paddingXY 10 0
                                    , Font.size 15
                                    , Font.color (Element.rgba255 0 0 0 0.7)
                                    ]
                                <|
                                    Element.text "Thread"
                            }
                )
                reply
            , Lib.maybe Element.none
                (\r ->
                    Element.el
                        [ Border.width 1
                        , Border.rounded 3
                        , Font.size 17
                        ]
                    <|
                        Element.el [ Element.padding 3 ] <|
                            viewPost True r.parent Nothing
                )
                reply
            , Element.link []
                { url = "https://bsky.app/profile/" ++ post.author.handle
                , label =
                    Element.row [ Element.spacing 5 ]
                        [ Element.text post.author.displayName
                        , Element.text <| "@" ++ post.author.handle
                        ]
                }
            , postText post.record
            , case post.embed of
                Just embed ->
                    embedView embed

                Nothing ->
                    Element.none
            , Element.link [ Font.size timestampSize ]
                { url = postUrl post
                , label =
                    Element.text <|
                        formatHMS ( 9, 0 ) <|
                            toHMS <|
                                post.record.createdAt
                }
            ]
        ]


viewFeed : Feed -> Element msg
viewFeed feed =
    Element.el
        [ Border.widthEach
            { bottom = 0
            , left = 0
            , right = 0
            , top = 1
            }
        , Element.paddingXY 0 10
        ]
    <|
        viewPost False feed.post feed.reply


view : Model -> Element msg
view model =
    Element.column
        [ Element.spacing 10
        , Element.paddingXY 30 0
        ]
        [ Element.row []
            [ Element.text "Date:"
            , Element.row []
                [ Element.text model.year
                , Element.text "/"
                , Element.text model.month
                , Element.text "/"
                , Element.text model.day
                ]
            ]
        , Element.column [] <| List.map viewFeed model.feeds
        ]
