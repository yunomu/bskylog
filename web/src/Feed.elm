module Feed exposing (Author, Facet, Feature, Feed, Index, Post, Record, decoder)

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


type Feature
    = Tag String
    | Link String


feature : Decoder Feature
feature =
    D.field "$type" D.string
        |> D.andThen
            (\t ->
                case t of
                    "app.bsky.richtext.facet#link" ->
                        D.map Link <| D.field "uri" D.string

                    "app.bsky.richtext.facet#tag" ->
                        D.map Tag <| D.field "tag" D.string

                    _ ->
                        D.fail <| "unknown tag: " ++ t
            )


type alias Index =
    { start : Int
    , end : Int
    }


index : Decoder Index
index =
    D.map2 Index
        (D.field "byteStart" D.int)
        (D.field "byteEnd" D.int)


type alias Facet =
    { features : List Feature
    , index : Index
    }


facet : Decoder Facet
facet =
    D.map2 Facet
        (D.field "features" <| D.list feature)
        (D.field "index" index)


type alias Record =
    { createdAt : String
    , facets : List Facet
    , text : String
    }


listField : String -> Decoder a -> Decoder (List a)
listField label d =
    (D.maybe <| D.field label <| D.list d)
        |> D.andThen (D.succeed << Maybe.withDefault [])


record : Decoder Record
record =
    D.map3 Record
        (D.field "createdAt" D.string)
        (listField "facets" facet)
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
