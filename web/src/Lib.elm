module Lib exposing (any, foldResult, maybe, perform)

import Task


maybe : b -> (a -> b) -> Maybe a -> b
maybe b f =
    Maybe.withDefault b << Maybe.map f


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


any : List Bool -> Bool
any ls =
    case ls of
        x :: xs ->
            if x then
                True

            else
                any xs

        [] ->
            False


perform : msg -> Cmd msg
perform =
    Task.perform identity << Task.succeed
