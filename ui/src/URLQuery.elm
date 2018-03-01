module URLQuery exposing (URLQuery, add, empty, render)

import Dict exposing (Dict)
import Http exposing (encodeUri)


type URLQuery
    = URLQuery (Dict String (List String))


empty : URLQuery
empty =
    URLQuery <| Dict.empty


add : String -> String -> URLQuery -> URLQuery
add key value (URLQuery query) =
    URLQuery <|
        Dict.update key
            (Maybe.withDefault []
                >> (::) value
                >> Just
            )
            query


render : URLQuery -> String
render (URLQuery query) =
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
