module Route exposing (Route(..), fromUrl, path, user)

import Url exposing (Url)
import Url.Builder as B
import Url.Parser as P exposing ((</>), (<?>), Parser)
import Url.Parser.Query as Q


type Route
    = Index
    | User String
    | Day String String String String
    | Search String (Maybe String)
    | NotFound Url


path : Route -> Maybe String
path route =
    case route of
        Index ->
            Just <| B.absolute [] []

        User u ->
            Just <| B.absolute [ u ] []

        Day u year month day ->
            Just <| B.absolute [ u, year, month, day ] []

        Search u mquery ->
            case mquery of
                Just query ->
                    Just <| B.absolute [ u, "search" ] [ B.string "q" query ]

                Nothing ->
                    Just <| B.absolute [ u, "search" ] []

        _ ->
            Nothing


user : Route -> Maybe String
user route =
    case route of
        User u ->
            Just u

        Day u _ _ _ ->
            Just u

        Search u _ ->
            Just u

        _ ->
            Nothing


parser : Parser (Route -> a) a
parser =
    P.oneOf
        [ P.map Index P.top
        , P.map User P.string
        , P.map Day <| P.string </> P.string </> P.string </> P.string
        , P.map (\u -> Search u Nothing) <| P.string </> P.s "search"
        , P.map Search <| P.string </> P.s "search" <?> Q.string "q"
        ]


fromUrl : Url -> Route
fromUrl url =
    Maybe.withDefault (NotFound url) <| P.parse parser url
