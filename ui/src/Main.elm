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
        [ view_header
        , Html.br [] []
        , Html.div [ Attr.class "container" ]
            [ view_centeredRow
                [ view_canvas
                ]
            , view_centeredRow
                [ view_inputForm model
                , view_errorText model
                ]
            ]
        ]


view_centeredRow : List (Html msg) -> Html msg
view_centeredRow children =
    Html.div [ Attr.class "row" ]
        [ Html.div [ Attr.class "col-sm-8 col-sm-offset-2" ] children
        ]


view_canvas : Html msg
view_canvas =
    Html.div [ Attr.id "canvas-container" ]
        [ Html.canvas [ Attr.id "canvas" ] [] ]


view_errorText : Model -> Html msg
view_errorText model =
    case model.graphState of
        Failed error ->
            Html.pre [] [ text error ]

        _ ->
            text ""


view_header : Html Msg
view_header =
    Html.header [ Attr.class "sticky" ]
        [ Html.a [ Attr.href "#", Attr.class "logo" ] [ text "Rothko" ]
        ]


view_inputForm : Model -> Html Msg
view_inputForm model =
    Html.form [ Attr.action "#" ]
        [ Html.fieldset []
            [ Html.legend [] [ text "Graph" ]
            , view_entry "metric" "goflud.env.process.uptime"
            , view_entry "duration" "24h"
            , view_entry "width" "1000"
            , view_entry "height" "360"
            , view_button "update" Draw
            ]
        ]


utils_titleCase : String -> String
utils_titleCase val =
    let
        upperWord word =
            String.toUpper (String.left 1 word) ++ (String.dropLeft 1 word)
    in
        val
            |> String.words
            |> List.map upperWord
            |> String.join " "


view_entry : String -> String -> Html msg
view_entry id value =
    Html.div [ Attr.class "row responsive-label" ]
        [ Html.div [ Attr.class "col-sm-12 col-md-3" ]
            [ Html.label [ Attr.for id ] [ text (utils_titleCase id) ]
            ]
        , Html.div [ Attr.class "col-sm-12 col-md-9" ]
            [ Html.input
                [ Attr.value value
                , Attr.id id
                , Attr.style [ ( "width", "100%" ) ]
                ]
                []
            ]
        ]


view_button : String -> msg -> Html msg
view_button id msg =
    Html.div [ Attr.class "row" ]
        [ Html.div [ Attr.class "col-sm-12 col-md-12" ]
            [ Html.button
                [ Attr.id id
                , Attr.class "primary"
                , Ev.onClick msg
                ]
                [ text (utils_titleCase id) ]
            ]
        ]
