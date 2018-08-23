module Api exposing
    ( RenderRequest
    , URLQueryRequest
    , doQuery
    , doRender
    , makeUrl
    , maybeAdd
    , queryRequest
    , queryRequestURLQuery
    , renderRequest
    , renderRequestURLQuery
    )

import Dict exposing (Dict)
import Http
import Json.Decode as Decode
import Url.Builder as Builder exposing (QueryParameter)



-- HELPERS


makeUrl : String -> List QueryParameter -> String
makeUrl path query =
    path ++ Builder.toQuery query


add : String -> String -> List QueryParameter -> List QueryParameter
add key val query =
    Builder.string key val :: query


maybeAdd : String -> Maybe String -> List QueryParameter -> List QueryParameter
maybeAdd key val query =
    case val of
        Nothing ->
            query

        Just val_ ->
            add key val_ query


maybeInt : Maybe Int -> Maybe String
maybeInt =
    Maybe.map String.fromInt


maybeFloat : Maybe Float -> Maybe String
maybeFloat =
    Maybe.map String.fromFloat



-- RENDER


type alias RenderRequest =
    { metric : String
    , width : Maybe Int
    , height : Maybe Int
    , padding : Maybe Int
    , now : Maybe Int
    , duration : Maybe String
    , samples : Maybe Int
    , compression : Maybe Float
    }


renderRequest : String -> RenderRequest
renderRequest metric =
    { metric = metric
    , width = Nothing
    , height = Nothing
    , padding = Nothing
    , now = Nothing
    , duration = Nothing
    , samples = Nothing
    , compression = Nothing
    }


doRender : RenderRequest -> Http.Request String
doRender req =
    Http.request
        { method = "GET"
        , headers = [ Http.header "Accept" "application/json" ]
        , url = makeUrl "/api/render" (renderRequestURLQuery req)
        , body = Http.emptyBody
        , expect = Http.expectString
        , timeout = Nothing
        , withCredentials = False
        }


renderRequestURLQuery : RenderRequest -> List QueryParameter
renderRequestURLQuery req =
    []
        |> add "metric" req.metric
        |> maybeAdd "width" (maybeInt req.width)
        |> maybeAdd "height" (maybeInt req.height)
        |> maybeAdd "padding" (maybeInt req.padding)
        |> maybeAdd "now" (maybeInt req.now)
        |> maybeAdd "duration" req.duration
        |> maybeAdd "samples" (maybeInt req.samples)
        |> maybeAdd "compression" (maybeFloat req.compression)


type alias URLQueryRequest =
    { query : String
    , results : Maybe Int
    }



-- QUERY


queryRequest : String -> URLQueryRequest
queryRequest query =
    { query = query
    , results = Nothing
    }


doQuery : URLQueryRequest -> Http.Request (List String)
doQuery req =
    Http.get (makeUrl "/api/query" (queryRequestURLQuery req)) (Decode.list Decode.string)


queryRequestURLQuery : URLQueryRequest -> List QueryParameter
queryRequestURLQuery req =
    []
        |> add "query" req.query
        |> maybeAdd "results" (maybeInt req.results)
