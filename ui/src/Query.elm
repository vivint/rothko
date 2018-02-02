module Query exposing (Query, empty, add, render)

import Http exposing (encodeUri)
import Dict exposing (Dict)


type Query
    = Query (Dict String (List String))


empty : Query
empty =
    Query <| Dict.empty


add : String -> String -> Query -> Query
add key value (Query query) =
    Query <|
        Dict.update key
            (Maybe.withDefault []
                >> (::) value
                >> Just
            )
            query


render : Query -> String
render (Query query) =
    let
        encode key value =
            encodeUri key ++ "=" ++ encodeUri value

        flatten ( key, values ) =
            List.map (encode key) values
    in
        Dict.toList query
            |> List.concatMap flatten
            |> String.join "&"
            |> (++) "?"
