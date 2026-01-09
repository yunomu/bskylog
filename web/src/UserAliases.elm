module UserAliases exposing (build)

import Csv.Builder as B exposing (Builder)
import Dict exposing (Dict)


builder : Builder ( String, String )
builder =
    B.fmap Tuple.pair
        B.string
        |> B.apply_ B.string


build : String -> Result String (Dict String String)
build =
    Result.map Dict.fromList << B.build builder
