module HttpLib exposing
    ( andThen
    , get
    )

import Http


get : (Result Http.Error String -> msg) -> String -> Cmd msg
get toMsg url =
    Http.get
        { url = url
        , expect = Http.expectString toMsg
        }


error : Http.Error -> String
error e =
    "HTTP ERROR"


andThen : (String -> Result String a) -> Result Http.Error String -> Result String a
andThen b res =
    Result.andThen b (Result.mapError error res)
