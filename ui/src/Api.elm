module Api exposing (..)

import Http
import URLQuery exposing (URLQuery)
import Dict exposing (Dict)
import Json.Decode exposing (list, string)


-- HELPERS


makeUrl : String -> URLQuery -> String
makeUrl path query =
    path ++ URLQuery.render query


maybeAdd : String -> Maybe String -> URLQuery -> URLQuery
maybeAdd key val query =
    case val of
        Nothing ->
            query

        Just val ->
            URLQuery.add key val query



-- RENDER


type alias RenderRequest =
    { metric : String
    , width : Maybe Int
    , height : Maybe Int
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
    , now = Nothing
    , duration = Nothing
    , samples = Nothing
    , compression = Nothing
    }


render : RenderRequest -> Http.Request String
render req =
    Http.request
        { method = "GET"
        , headers = [ Http.header "Accept" "application/json" ]
        , url = makeUrl "/api/render" (renderRequestURLQuery req)
        , body = Http.emptyBody
        , expect = Http.expectString
        , timeout = Nothing
        , withCredentials = False
        }


renderRequestURLQuery : RenderRequest -> URLQuery
renderRequestURLQuery req =
    let
        ts =
            Maybe.map toString
    in
        URLQuery.empty
            |> URLQuery.add "metric" req.metric
            |> maybeAdd "width" (ts req.width)
            |> maybeAdd "height" (ts req.height)
            |> maybeAdd "now" (ts req.now)
            |> maybeAdd "duration" req.duration
            |> maybeAdd "samples" (ts req.samples)
            |> maybeAdd "compression" (ts req.compression)


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


query : URLQueryRequest -> Http.Request (List String)
query req =
    Http.get (makeUrl "/api/query" (queryRequestURLQuery req)) (list string)


queryRequestURLQuery : URLQueryRequest -> URLQuery
queryRequestURLQuery req =
    let
        ts =
            Maybe.map toString
    in
        URLQuery.empty
            |> URLQuery.add "query" req.query
            |> maybeAdd "results" (ts req.results)
