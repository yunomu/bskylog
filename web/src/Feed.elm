module Feed exposing
    ( Author
    , Embed(..)
    , External
    , Facet
    , Feature(..)
    , Feed
    , Image
    , Index
    , Post
    , Record
    , Reply
    , Video
    , decoder
    )

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


type alias Size =
    { width : Int
    , height : Int
    }


size : Decoder Size
size =
    D.map2 Size
        (D.field "width" D.int)
        (D.field "height" D.int)


type alias Image =
    { alt : String
    , aspectRatio : Size
    , fullsize : String
    , thumb : String
    }


image : Decoder Image
image =
    D.map4 Image
        (D.field "alt" D.string)
        (D.field "aspectRatio" size)
        (D.field "fullsize" D.string)
        (D.field "thumb" D.string)


type alias External =
    { description : String
    , thumb : String
    , title : String
    , uri : String
    }


external : Decoder External
external =
    D.map4 External
        (D.field "description" D.string)
        (D.field "thumb" D.string)
        (D.field "title" D.string)
        (D.field "uri" D.string)


type alias Video =
    { aspectRatio : Size
    , playlist : String
    , thumbnail : String
    }


video : Decoder Video
video =
    D.map3 Video
        (D.field "aspectRatio" size)
        (D.field "playlist" D.string)
        (D.field "thumbnail" D.string)


type Embed
    = EmbeddedImage (List Image)
    | EmbeddedVideo Video
    | EmbeddedExternal External


embed : Decoder Embed
embed =
    D.field "$type" D.string
        |> D.andThen
            (\t ->
                case t of
                    "app.bsky.embed.images#view" ->
                        D.map EmbeddedImage <| D.field "images" <| D.list image

                    "app.bsky.embed.video#view" ->
                        D.map EmbeddedVideo video

                    "app.bsky.embed.external#view" ->
                        D.map EmbeddedExternal <| D.field "external" external

                    _ ->
                        D.fail <| "unknown embed type: " ++ t
            )


type alias Post =
    { author : Author
    , cid : String
    , embed : Maybe Embed
    , record : Record
    , uri : String
    }


post : Decoder Post
post =
    D.map5 Post
        (D.field "author" author)
        (D.field "cid" D.string)
        (D.maybe <| D.field "embed" embed)
        (D.field "record" record)
        (D.field "uri" D.string)


type alias Reply =
    { parent : Post
    , root : Post
    }


reply : Decoder Reply
reply =
    D.map2 Reply
        (D.field "parent" post)
        (D.field "root" post)


type alias Feed =
    { post : Post
    , reply : Maybe Reply
    }


decoder : Decoder Feed
decoder =
    D.map2 Feed
        (D.field "post" post)
        (D.maybe <| D.field "reply" reply)
