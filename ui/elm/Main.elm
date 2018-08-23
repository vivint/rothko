module Main exposing
    ( Flags
    , Model
    , Msg(..)
    , Path
    , graphConfig
    , init
    , main
    , parse
    , queryBarConfig
    , render
    , sendMessage
    , subscriptions
    , update
    , view
    , viewFullRow
    , viewGraph
    , viewHeader
    , viewQuery
    )

import Api
import Browser
import Browser.Navigation as Navigation
import Graph
import Html exposing (Html, text)
import Html.Attributes as Attr
import Http
import Json.Decode as Decode
import Json.Encode as Encode exposing (Value)
import Process
import QueryBar
import Task
import Url exposing (Url)
import Url.Builder as Builder exposing (QueryParameter)
import Url.Parser as Parser exposing ((<?>))
import Url.Parser.Query as Query



-- MAIN


main =
    Browser.application
        { init = init
        , view = view
        , update = update
        , subscriptions = subscriptions
        , onUrlRequest = OnUrlRequest
        , onUrlChange = OnUrlChange
        }


type alias Flags =
    {}



-- MSG


type Msg
    = Draw String Bool
    | GraphMsg Graph.Msg
    | QueryBarMsg QueryBar.Msg
    | OnUrlRequest Browser.UrlRequest
    | OnUrlChange Url



-- PATH


type alias Path =
    { metric : Maybe String
    }


parse : Url -> Path
parse url =
    let
        parser =
            Parser.top <?> Query.string "metric"
    in
    Parser.parse (Parser.map Path parser) url
        |> Maybe.withDefault { metric = Nothing }


render : Path -> String
render { metric } =
    let
        maybeAdd : String -> Maybe String -> List QueryParameter -> List QueryParameter
        maybeAdd key val params =
            case val of
                Nothing ->
                    params

                Just val_ ->
                    Builder.string key val_ :: params
    in
    []
        |> maybeAdd "metric" metric
        |> Builder.toQuery



-- MODEL


type alias Model =
    { flags : Flags
    , key : Navigation.Key
    , url : Url
    , graph : Graph.Model
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
    , select = \metric -> Draw metric True
    }



-- INIT


init : Flags -> Url -> Navigation.Key -> ( Model, Cmd Msg )
init flags url key =
    let
        metric =
            parse url |> .metric |> Maybe.withDefault ""

        ( graph, cmd ) =
            Graph.new graphConfig
    in
    ( { flags = flags
      , key = key
      , url = url
      , graph = graph
      , queryBar = QueryBar.new metric
      }
    , Cmd.batch
        [ cmd
        , if metric /= "" then
            sendMessage <| Draw metric False

          else
            Cmd.none
        ]
    )



-- UPDATE


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        OnUrlChange url ->
            if url /= model.url then
                init model.flags url model.key

            else
                ( model, Cmd.none )

        OnUrlRequest request ->
            case request of
                Browser.Internal url ->
                    ( model, Navigation.load (Url.toString url) )

                Browser.External url ->
                    ( model, Navigation.load url )

        Draw metric push ->
            ( model
            , Cmd.batch
                [ sendMessage <| GraphMsg <| Graph.Draw metric
                , if push then
                    Navigation.pushUrl model.key (render { metric = Just metric })

                  else
                    Cmd.none
                ]
            )

        GraphMsg graphMsg ->
            Graph.update graphConfig model graphMsg

        QueryBarMsg queryBarMsg ->
            QueryBar.update queryBarConfig model queryBarMsg



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ Graph.subscriptions graphConfig model
        , QueryBar.subscriptions queryBarConfig model
        ]



-- VIEW


view : Model -> Browser.Document Msg
view model =
    { title = "Rothko"
    , body =
        [ viewHeader
        , Html.div [ Attr.class "container" ]
            [ viewFullRow [ viewQuery model ]
            , viewFullRow [ viewGraph model ]
            ]
        ]
    }


viewHeader : Html Msg
viewHeader =
    Html.header [ Attr.class "sticky" ]
        [ Html.a [ Attr.href "#", Attr.class "logo" ] [ text "Rothko" ]
        ]


viewFullRow : List (Html msg) -> Html msg
viewFullRow children =
    Html.div [ Attr.class "row", Attr.style "padding-top" "10px" ]
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
