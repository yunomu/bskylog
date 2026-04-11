module View.Day exposing
    ( Model
    , Msg(..)
    , init
    , update
    , view
    )

import Element exposing (Element)
import Element.Lazy as Lazy
import Feed exposing (Feed)
import Lib
import Task
import View.Org.Feeds


type alias Model msg =
    { user : String
    , year : String
    , month : String
    , day : String
    , feeds : List Feed
    , changed : String -> String -> String -> String -> msg
    }


init : (String -> String -> String -> String -> msg) -> Model msg
init changed =
    Model "" "" "" "" [] changed


type Msg
    = UpdateDay String String String String
    | UpdateFeeds (List Feed)


update : Msg -> Model msg -> ( Model msg, Cmd msg )
update msg model =
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
                , Lib.perform <| model.changed user year month day
                )

            else
                ( model, Cmd.none )

        UpdateFeeds feeds ->
            ( { model | feeds = feeds }, Cmd.none )


view_ : Model msg -> Element msg
view_ model =
    Element.column
        [ Element.alignTop
        , Element.spacing 10
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
        , View.Org.Feeds.view model.feeds
        ]


view : Element msg -> Model msg -> Element msg
view side model =
    Element.row []
        [ Lazy.lazy view_ model
        , side
        ]
