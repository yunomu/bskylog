module Route exposing (Route(..), fromUrl, path)

import Url exposing (Url)
import Url.Builder as B
import Url.Parser as P exposing ((</>), Parser)


type Route
    = Index
    | User String
    | UserDate String String String String
    | NotFound Url


path : Route -> Maybe String
path route =
    case route of
        Index ->
            Just <| B.absolute [] []

        User user ->
            Just <| B.absolute [ user ] []

        UserDate user year month day ->
            Just <| B.absolute [ user, year, month, day ] []

        _ ->
            Nothing


parser : Parser (Route -> a) a
parser =
    P.oneOf
        [ P.map Index P.top
        , P.map User P.string
        , P.map UserDate <| P.string </> P.string </> P.string </> P.string
        ]


fromUrl : Url -> Route
fromUrl url =
    Maybe.withDefault (NotFound url) <| P.parse parser url
