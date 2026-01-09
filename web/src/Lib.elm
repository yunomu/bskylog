module Lib exposing (foldResult)


cons : a -> List a -> List a
cons x xs =
    x :: xs


foldResult : List (Result x a) -> Result x (List a)
foldResult xs =
    case xs of
        r :: rs ->
            case r of
                Ok a ->
                    Result.map (cons a) (foldResult rs)

                Err err ->
                    Err err

        [] ->
            Ok []
