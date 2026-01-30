module View.Org.Month exposing
    ( Model
    , Msg(..)
    , init
    , update
    , view
    )

import Element exposing (Element)
import Lib
import MonthIndex exposing (Day)


type alias Model msg =
    { user : String
    , year : String
    , month : String
    , index : List Day
    , changed : String -> String -> String -> msg
    }


init : (String -> String -> String -> msg) -> Model msg
init changed =
    Model "" "" "" [] changed


type Msg
    = UpdateMonth String String String
    | UpdateIndex (List Day)


update : Msg -> Model msg -> ( Model msg, Cmd msg )
update msg model =
    case msg of
        UpdateMonth user year month ->
            if
                Lib.any
                    [ user /= model.user
                    , year /= model.year
                    , month /= model.month
                    ]
            then
                ( { model
                    | user = user
                    , year = year
                    , month = month
                  }
                , Lib.perform <| model.changed user year month
                )

            else
                ( model, Cmd.none )

        UpdateIndex days ->
            ( { model | index = days }
            , Cmd.none
            )


d02 : Int -> String
d02 i =
    let
        s =
            String.fromInt i
    in
    if String.length s == 1 then
        "0" ++ s

    else
        s


view : Model msg -> Element msg
view model =
    Element.column
        [ Element.alignTop
        , Element.spacing 10
        ]
        [ Element.text <|
            String.concat
                [ model.year
                , "/"
                , model.month
                ]
        , Element.column
            [ Element.paddingXY 20 0
            , Element.spacing 3
            ]
          <|
            List.map
                (\d ->
                    Element.link []
                        { url = d.day
                        , label =
                            Element.text <|
                                String.concat
                                    [ d.day
                                    , " ("
                                    , d02 d.count
                                    , ")"
                                    ]
                        }
                )
                model.index
        ]
