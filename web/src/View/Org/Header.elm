module View.Org.Header exposing (view)

import Element exposing (Element)
import Element.Border as Border


edges =
    { top = 0
    , bottom = 0
    , left = 0
    , right = 0
    }


view : Element msg
view =
    Element.row
        [ Element.width Element.fill
        , Element.padding 5
        , Border.widthEach { edges | bottom = 1 }
        ]
        [ Element.link [ Element.alignLeft ]
            { url = "/"
            , label = Element.text "BskyLog"
            }
        ]
