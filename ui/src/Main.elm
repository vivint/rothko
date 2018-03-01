module Main exposing (..)

import Api
import Graph
import Html exposing (Html, text)
import Html.Attributes as Attr
import Http
import Json.Decode as Decode
import Json.Encode as Encode exposing (Value)
import Navigation
import Process
import QueryBar
import Task
import Time exposing (Time)
import URLQuery exposing (URLQuery)
import UrlParser exposing ((<?>), map, parsePath, stringParam, top)
import Window


-- MAIN


main =
    Navigation.programWithFlags Location
        { init = init
        , update = update
        , subscriptions = subscriptions
        , view = view
        }


type alias Flags =
    {}



-- MSG


type Msg
    = Location Navigation.Location
    | Draw String
    | GraphMsg Graph.Msg
    | QueryBarMsg QueryBar.Msg



-- PATH


type alias Path =
    { metric : Maybe String
    }


parse : Navigation.Location -> Path
parse loc =
    let
        parser =
            top
                <?> stringParam "metric"
    in
    parsePath (map Path parser) loc
        |> Maybe.withDefault
            { metric = Nothing
            }


render : Path -> String
render { metric } =
    let
        maybeAdd : String -> Maybe String -> URLQuery -> URLQuery
        maybeAdd key val query =
            case val of
                Nothing ->
                    query

                Just val ->
                    URLQuery.add key val query
    in
    URLQuery.empty
        |> maybeAdd "metric" metric
        |> URLQuery.render



-- MODEL


type alias Model =
    { graph : Graph.Model
    , queryBar : QueryBar.Model
    }


graphConfig : Graph.Config Model Msg
graphConfig =
    { get = .graph
    , set = \model graph -> { model | graph = graph }
    , wrap = GraphMsg
    }


queryBarConfig : QueryBar.Config Model Msg
queryBarConfig =
    { get = .queryBar
    , set = \model queryBar -> { model | queryBar = queryBar }
    , wrap = QueryBarMsg
    , select = Draw
    }



-- INIT


init : Flags -> Navigation.Location -> ( Model, Cmd Msg )
init flags loc =
    let
        metric =
            parse loc |> .metric |> Maybe.withDefault ""

        ( graph, cmd ) =
            Graph.new graphConfig
    in
    ( { graph = graph
      , queryBar = QueryBar.new metric
      }
    , Cmd.batch
        [ cmd
        , if metric /= "" then
            sendMessage <| Draw metric
          else
            Cmd.none
        ]
    )



-- UPDATE


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        Location loc ->
            ( model, Cmd.none )

        Draw metric ->
            ( model
            , Cmd.batch
                [ sendMessage <| GraphMsg <| Graph.Draw metric
                , Navigation.newUrl (render { metric = Just metric })
                ]
            )

        GraphMsg msg ->
            Graph.update graphConfig model msg

        QueryBarMsg msg ->
            QueryBar.update queryBarConfig model msg



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ Graph.subscriptions graphConfig model
        , QueryBar.subscriptions queryBarConfig model
        ]



-- VIEW


view : Model -> Html Msg
view model =
    Html.div []
        [ viewHeader
        , Html.div [ Attr.class "container" ]
            [ viewFullRow [ viewQuery model ]
            , viewFullRow [ viewGraph model ]
            ]
        ]


viewHeader : Html Msg
viewHeader =
    Html.header [ Attr.class "sticky" ]
        [ Html.a [ Attr.href "#", Attr.class "logo" ] [ text "Rothko" ]
        ]


viewFullRow : List (Html msg) -> Html msg
viewFullRow children =
    Html.div [ Attr.class "row", Attr.style [ ( "padding-top", "10px" ) ] ]
        [ Html.div [ Attr.class "col-sm-12" ] children
        ]


viewQuery : Model -> Html Msg
viewQuery model =
    QueryBar.view queryBarConfig model.queryBar


viewGraph : Model -> Html Msg
viewGraph model =
    Graph.view graphConfig model.graph



-- UTILS


sendMessage : msg -> Cmd msg
sendMessage msg =
    Task.perform (always msg) (Task.succeed ())
