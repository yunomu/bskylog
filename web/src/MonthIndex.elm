module MonthIndex exposing (Day, build)

import Csv.Builder as B exposing (Builder)
import Dict exposing (Dict)


type alias Day =
    { day : String
    , count : Int
    }


builder : Builder Day
builder =
    B.fmap Day
        B.string
        |> B.apply_ B.int


build : String -> Result String (List Day)
build =
    B.build builder
