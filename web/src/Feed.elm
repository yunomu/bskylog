module Feed exposing (Author, Feed, Post, Record, decoder)

import Json.Decode as D exposing (Decoder, Value)


type alias Author =
    { avatar : String
    , did : String
    , displayName : String
    , handle : String
    }


author : Decoder Author
author =
    D.map4 Author
        (D.field "avatar" D.string)
        (D.field "did" D.string)
        (D.field "displayName" D.string)
        (D.field "handle" D.string)


type alias Record =
    { createdAt : String
    , text : String
    }


record : Decoder Record
record =
    D.map2 Record
        (D.field "createdAt" D.string)
        (D.field "text" D.string)


type alias Post =
    { author : Author
    , cid : String
    , record : Record
    , uri : String
    }


type alias Feed =
    { post : Post
    }


decoder : Decoder Feed
decoder =
    D.map Feed <|
        D.field "post" <|
            D.map4 Post
                (D.field "author" author)
                (D.field "cid" D.string)
                (D.field "record" record)
                (D.field "uri" D.string)
