module Main exposing (..)

import Api
import Graph
import QueryBar
import Html exposing (Html, text)
import Html.Attributes as Attr
import Http


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
    = NoOp
    | DrawStarting
    | DrawDone Graph.Result
    | RenderResponse (Result Http.Error String)
    | QueryBarMsg QueryBar.Msg
    | Select String



-- MODEL


type GraphState
    = Stale
    | Drawing
    | Fresh
    | Failed String


type alias Model =
    { graphState : GraphState
    , queryBar : QueryBar.Model
    }


queryBarConfig : QueryBar.Config Model Msg
queryBarConfig =
    { get = .queryBar
    , set = \model queryBar -> { model | queryBar = queryBar }
    , wrap = QueryBarMsg
    , select = Select
    }



-- INIT


init : Flags -> ( Model, Cmd Msg )
init flags =
    ( { graphState = Stale
      , queryBar = QueryBar.new
      }
    , Cmd.none
    )



-- UPDATE


withModel : Model -> Cmd Msg -> ( Model, Cmd Msg )
withModel model msg =
    ( model, msg )


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        NoOp ->
            ( model, Cmd.none )

        DrawStarting ->
            ( { model | graphState = Drawing }, Cmd.none )

        DrawDone result ->
            ( if result.ok then
                { model | graphState = Fresh }
              else
                { model | graphState = Failed result.error }
            , Cmd.none
            )

        RenderResponse response ->
            case response of
                Ok result ->
                    ( model, Graph.draw result )

                Err error ->
                    Debug.log (toString error) ( model, Cmd.none )

        Select value ->
            ( model
            , Api.renderRequest value
                |> Api.render
                |> Http.send RenderResponse
            )

        QueryBarMsg msg ->
            QueryBar.update queryBarConfig model msg



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ Graph.starting DrawStarting
        , Graph.done DrawDone
        , QueryBar.subscriptions queryBarConfig model
        ]



-- VIEW


view : Model -> Html Msg
view model =
    Html.div []
        [ view_header
        , Html.br [] []
        , Html.div [ Attr.class "container" ]
            [ view_fullRow
                [ view_canvas
                ]
            , view_fullRow
                [ view_query model
                , view_errorText model
                ]
            ]
        ]


view_fullRow : List (Html msg) -> Html msg
view_fullRow children =
    Html.div [ Attr.class "row" ]
        [ Html.div [ Attr.class "col-sm-12" ] children
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


view_query : Model -> Html Msg
view_query model =
    view_fullRow [ QueryBar.view queryBarConfig model.queryBar ]
