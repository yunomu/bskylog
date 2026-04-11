module View.Search exposing
    ( Model
    , Msg(..)
    , init
    , update
    , view
    )

import Element exposing (Element)
import Element.Input as Input
import Element.Lazy as Lazy
import Feed exposing (Feed)
import Lib
import Task
import View.Atom.Button
import View.Org.Feeds


type alias Model msg =
    { query : String
    , feeds : List Feed
    , submit : String -> msg
    }


init : (String -> msg) -> Model msg
init submit =
    Model "" [] submit


type Msg
    = UpdateFeeds (List Feed)
    | QueryChanged String
    | Submit
    | ClearQuery


update : Msg -> Model msg -> ( Model msg, Cmd msg )
update msg model =
    case msg of
        UpdateFeeds feeds ->
            ( { model | feeds = feeds }, Cmd.none )

        QueryChanged query ->
            ( { model | query = query }
            , Cmd.none
            )

        Submit ->
            ( model
            , if String.isEmpty model.query then
                Cmd.none

              else
                Lib.perform <| model.submit model.query
            )

        ClearQuery ->
            ( { model | query = "" }
            , Cmd.none
            )


view : (Msg -> msg) -> Model msg -> Element msg
view toMsg model =
    Element.column
        [ Element.alignTop
        , Element.spacing 10
        , Element.paddingXY 30 0
        ]
        [ Element.row
            [ Element.spacing 10
            ]
            [ Input.text []
                { onChange = toMsg << QueryChanged
                , text = model.query
                , placeholder = Nothing
                , label = Input.labelLeft [] <| Element.text "Query:"
                }
            , View.Atom.Button.updateButton (toMsg Submit) "Submit"
            , View.Atom.Button.button (toMsg ClearQuery) "Clear"
            ]
        , View.Org.Feeds.view model.feeds
        ]
