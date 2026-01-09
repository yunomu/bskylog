module View exposing (template)

import Element exposing (Element)


template : Element msg -> Element msg -> Element msg
template header content =
    Element.column
        [ Element.centerX
        , Element.width Element.fill
        , Element.padding 5
        , Element.spacing 20
        ]
        [ header
        , content
        ]
