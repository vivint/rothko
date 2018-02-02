module Api exposing (..)

import Http
import Query exposing (Query)
import Dict exposing (Dict)
import Json.Decode exposing (list, string)


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
        , url = "http://localhost:9998/render" ++ renderRequestQuery req
        , body = Http.emptyBody
        , expect = Http.expectString
        , timeout = Nothing
        , withCredentials = False
        }


maybeAdd : String -> Maybe a -> Query -> Query
maybeAdd key val query =
    case val of
        Nothing ->
            query

        Just val ->
            Query.add key (Basics.toString val) query


renderRequestQuery : RenderRequest -> String
renderRequestQuery req =
    Query.empty
        |> Query.add "metric" req.metric
        |> maybeAdd "width" req.width
        |> maybeAdd "height" req.height
        |> maybeAdd "now" req.now
        |> maybeAdd "duration" req.duration
        |> maybeAdd "samples" req.samples
        |> maybeAdd "compression" req.compression
        |> Query.render


type alias QueryRequest =
    { query : String
    , results : Maybe Int
    }


queryRequest : String -> QueryRequest
queryRequest query =
    { query = query
    , results = Nothing
    }


query : QueryRequest -> Http.Request (List String)
query req =
    let
        url =
            "http://localhost:9998/query" ++ queryRequestQuery req
    in
        Http.get url (list string)


queryRequestQuery : QueryRequest -> String
queryRequestQuery req =
    Query.empty
        |> Query.add "query" req.query
        |> maybeAdd "results" req.results
        |> Query.render
