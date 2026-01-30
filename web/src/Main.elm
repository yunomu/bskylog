module Main exposing (main)

import Browser
import Browser.Events as Events
import Browser.Navigation as Nav
import Dict exposing (Dict)
import Element exposing (Element)
import Element.Lazy as Lazy
import Feed exposing (Feed)
import Http
import HttpLib
import Json.Decode as JD
import Lib
import MonthIndex
import Route exposing (Route)
import Task
import Url exposing (Url)
import Url.Builder as UrlBuilder
import UserAliases
import View
import View.Day
import View.Index
import View.Org.Header
import View.Org.Month


type alias Flags =
    { windowWidth : Int
    , windowHeight : Int
    }


type Msg
    = UrlRequest Browser.UrlRequest
    | UrlChanged Url
    | OnResize Int Int
    | InitFetchUserCsv (Cmd Msg) (Result Http.Error String)
    | FetchLog (Result Http.Error String)
    | FetchMonthIndex (Result Http.Error String)
    | DayMsg View.Day.Msg
    | MonthMsg View.Org.Month.Msg
    | MonthChanged String String String


type alias Model =
    { key : Nav.Key
    , route : Route
    , windowSize : ( Int, Int )
    , userAliases : Dict String String
    , dayModel : View.Day.Model
    , monthModel : View.Org.Month.Model Msg
    }


init : Flags -> Url -> Nav.Key -> ( Model, Cmd Msg )
init flags url key =
    ( { key = key
      , route = Route.fromUrl url
      , windowSize = ( flags.windowWidth, flags.windowHeight )
      , userAliases = Dict.empty
      , dayModel = View.Day.init
      , monthModel = View.Org.Month.init MonthChanged
      }
    , Cmd.batch
        [ HttpLib.get (InitFetchUserCsv (Nav.pushUrl key (Url.toString url))) "/users"
        ]
    )


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        UrlRequest urlRequest ->
            case urlRequest of
                Browser.Internal url ->
                    ( model, Nav.pushUrl model.key (Url.toString url) )

                Browser.External href ->
                    ( model, Nav.load href )

        UrlChanged url ->
            let
                route =
                    Route.fromUrl url
            in
            case route of
                Route.Day user year month day ->
                    case Dict.get user model.userAliases of
                        Just did ->
                            ( { model | route = route }
                            , Cmd.batch
                                [ Lib.perform DayMsg <| View.Day.UpdateDay user year month day
                                , Lib.perform MonthMsg <| View.Org.Month.UpdateMonth user year month
                                ]
                            )

                        Nothing ->
                            ( { model | route = route }
                            , Nav.pushUrl model.key "/"
                            )

                _ ->
                    ( { model | route = route }
                    , Cmd.none
                    )

        OnResize w h ->
            ( { model | windowSize = ( w, h ) }
            , Cmd.none
            )

        InitFetchUserCsv cmd res ->
            case HttpLib.andThen UserAliases.build res of
                Ok userAliases ->
                    ( { model | userAliases = userAliases }, cmd )

                Err error ->
                    ( model, cmd )

        FetchLog res ->
            let
                decodeLog =
                    Result.mapError JD.errorToString
                        << Lib.foldResult
                        << List.map (JD.decodeString Feed.decoder)
                        << List.filter (not << String.isEmpty)
                        << String.lines
            in
            case HttpLib.andThen decodeLog res of
                Ok feeds ->
                    let
                        ( dayModel, dayCmd ) =
                            View.Day.update DayMsg (View.Day.UpdateFeeds feeds) model.dayModel
                    in
                    ( { model
                        | dayModel = dayModel
                      }
                    , dayCmd
                    )

                Err err ->
                    ( model, Cmd.none )

        DayMsg dayMsg ->
            let
                ( dayModel, dayCmd ) =
                    View.Day.update DayMsg dayMsg model.dayModel
            in
            case dayMsg of
                View.Day.Changed user year month day ->
                    Lib.maybe ( model, Cmd.none )
                        (\did ->
                            ( { model | dayModel = dayModel }
                            , HttpLib.get FetchLog <| UrlBuilder.absolute [ did, year, month, day ] []
                            )
                        )
                    <|
                        Dict.get user model.userAliases

                _ ->
                    ( { model | dayModel = dayModel }
                    , dayCmd
                    )

        MonthMsg monthMsg ->
            let
                ( mModel, mCmd ) =
                    View.Org.Month.update monthMsg model.monthModel
            in
            ( { model | monthModel = mModel }
            , mCmd
            )

        MonthChanged user year month ->
            Lib.maybe ( model, Cmd.none )
                (\did ->
                    ( model
                    , HttpLib.get FetchMonthIndex <| UrlBuilder.absolute [ did, year, month, "index" ] []
                    )
                )
            <|
                Dict.get user model.userAliases

        FetchMonthIndex res ->
            case HttpLib.andThen MonthIndex.build res of
                Ok days ->
                    let
                        ( mModel, mCmd ) =
                            View.Org.Month.update (View.Org.Month.UpdateIndex days) model.monthModel
                    in
                    ( { model | monthModel = mModel }
                    , mCmd
                    )

                Err error ->
                    ( model, Cmd.none )


subscriptions : Model -> Sub Msg
subscriptions _ =
    Events.onResize OnResize


view : Model -> Browser.Document Msg
view model =
    { title =
        case model.route of
            Route.Day user year month day ->
                String.concat
                    [ "Bskylog: "
                    , user
                    , " ("
                    , year
                    , "/"
                    , month
                    , "/"
                    , day
                    , ")"
                    ]

            _ ->
                "Bskylog"
    , body =
        [ Element.layout [] <|
            View.template
                View.Org.Header.view
                (case model.route of
                    Route.Index ->
                        View.Index.view

                    Route.Day user year month day ->
                        Lazy.lazy2 View.Day.view
                            (Lazy.lazy View.Org.Month.view model.monthModel)
                            model.dayModel

                    _ ->
                        View.Index.view
                )
        ]
    }


main : Program Flags Model Msg
main =
    Browser.application
        { init = init
        , update = update
        , subscriptions = subscriptions
        , view = view
        , onUrlChange = UrlChanged
        , onUrlRequest = UrlRequest
        }
