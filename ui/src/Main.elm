module Main exposing (..)

import Api
import Graph
import QueryBar
import Html exposing (Html, text)
import Html.Attributes as Attr
import Http
import Task
import Process
import Time exposing (Time)
import Window
import Json.Encode as Encode exposing (Value)
import Json.Decode as Decode


-- MAIN


main =
    Html.programWithFlags
        { init = init
        , update = update
        , subscriptions = subscriptions
        , view = view
        }


type alias Flags =
    {}



-- MSG


type Msg
    = Draw String
    | GraphMsg Graph.Msg
    | QueryBarMsg QueryBar.Msg



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


init : Flags -> ( Model, Cmd Msg )
init flags =
    let
        ( graph, cmd ) =
            Graph.new graphConfig
    in
        ( { graph = graph
          , queryBar = QueryBar.new
          }
        , cmd
        )



-- UPDATE


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        Draw metric ->
            ( model, sendMessage <| GraphMsg <| Graph.Draw metric )

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
