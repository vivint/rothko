module Main exposing (..)

import Graph
import Html exposing (Html, text)
import Html.Attributes as Attr
import Html.Events as Ev
import Window


main =
    Html.programWithFlags
        { init = init
        , update = update
        , subscriptions = subscriptions
        , view = view
        }


type alias Flags =
    {}


type Msg
    = DrawStarting
    | DrawDone Graph.Result
    | Draw


type GraphState
    = Stale
    | Drawing
    | Fresh
    | Failed String


type alias Model =
    { graphState : GraphState }


init : Flags -> ( Model, Cmd Msg )
init flags =
    ( { graphState = Stale
      }
    , Graph.draw
    )


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        DrawStarting ->
            ( { model | graphState = Drawing }, Cmd.none )

        DrawDone result ->
            ( if result.ok then
                { model | graphState = Fresh }
              else
                { model | graphState = Failed result.error }
            , Cmd.none
            )

        Draw ->
            ( model, Graph.draw )


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ Graph.starting DrawStarting
        , Graph.done DrawDone
        ]


view : Model -> Html Msg
view model =
    Html.div []
        [ header
        , Html.br [] []
        , Html.div [ Attr.class "container" ]
            [ Html.div [ Attr.class "row" ]
                [ Html.div [ Attr.class "col-sm-8 col-sm-offset-2", Attr.id "canvas-container" ]
                    [ Html.canvas [ Attr.id "canvas" ] []
                    , inputForm model
                    , case model.graphState of
                        Failed error ->
                            Html.pre [] [ text error ]

                        _ ->
                            text ""
                    ]
                ]
            ]
        ]


header : Html Msg
header =
    Html.header [ Attr.class "sticky" ]
        [ Html.a [ Attr.href "#", Attr.class "logo" ] [ text "Rothko" ]
        ]


inputForm : Model -> Html Msg
inputForm model =
    Html.form []
        [ Html.fieldset []
            [ Html.legend [] [ text "Graph" ]
            , Html.div [ Attr.class "row responsive-label" ]
                [ Html.div [ Attr.class "col-sm-12 col-md-3" ]
                    [ Html.label [ Attr.for "metric" ] [ text "Metric" ]
                    ]
                , Html.div [ Attr.class "col-sm-12 col-md-9" ]
                    [ Html.input [ Attr.value "goflud.env.process.uptime", Attr.id "metric", Attr.style [ ( "width", "100%" ) ] ] []
                    ]
                ]
            , Html.div [ Attr.class "row responsive-label" ]
                [ Html.div [ Attr.class "col-sm-12 col-md-3" ]
                    [ Html.label [ Attr.for "duration" ] [ text "Duration" ]
                    ]
                , Html.div [ Attr.class "col-sm-12 col-md-9" ]
                    [ Html.input [ Attr.value "24h", Attr.id "duration", Attr.style [ ( "width", "100%" ) ] ] []
                    ]
                ]
            , Html.div [ Attr.class "row" ]
                [ Html.div [ Attr.class "col-sm-12 col-md-12" ]
                    [ Html.button [ Attr.id "update", Attr.class "primary", Ev.onClick Draw ] [ text "Update" ] ]
                ]
            , Html.hr [] []
            , Html.div [ Attr.class "row responsive-label" ]
                [ Html.div [ Attr.class "col-sm-12 col-md-3" ]
                    [ Html.label [ Attr.for "width" ] [ text "Width" ]
                    ]
                , Html.div [ Attr.class "col-sm-12 col-md-9" ]
                    [ Html.input [ Attr.value "1000", Attr.id "width", Attr.style [ ( "width", "100%" ) ] ] []
                    ]
                ]
            , Html.div [ Attr.class "row responsive-label" ]
                [ Html.div [ Attr.class "col-sm-12 col-md-3" ]
                    [ Html.label [ Attr.for "height" ] [ text "Height" ]
                    ]
                , Html.div [ Attr.class "col-sm-12 col-md-9" ]
                    [ Html.input [ Attr.value "300", Attr.id "height", Attr.style [ ( "width", "100%" ) ] ] []
                    ]
                ]
            ]
        ]
