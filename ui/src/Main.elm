module Main exposing (..)

import Api
import Graph
import QueryBar
import Html exposing (Html, text)
import Html.Attributes as Attr
import Http
import Task
import Process
import Time
import Window


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
    | Resize Window.Size
    | ResizeTimer Window.Size
    | DrawStarting String
    | DrawDone Graph.Result
    | Render String
    | RenderResponse (Result Http.Error String)
    | QueryBarMsg QueryBar.Msg



-- MODEL


type GraphState
    = Stale
    | Drawing String
    | Fresh String
    | Failed String


isFresh : GraphState -> Bool
isFresh graphState =
    case graphState of
        Fresh _ ->
            True

        _ ->
            False


type alias Model =
    { graphState : GraphState
    , queryBar : QueryBar.Model
    , size : Maybe Window.Size
    }


queryBarConfig : QueryBar.Config Model Msg
queryBarConfig =
    { get = .queryBar
    , set = \model queryBar -> { model | queryBar = queryBar }
    , wrap = QueryBarMsg
    , select = Render
    }



-- INIT


init : Flags -> ( Model, Cmd Msg )
init flags =
    ( { graphState = Stale
      , queryBar = QueryBar.new
      , size = Nothing
      }
    , Task.perform Resize Window.size
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

        Resize size ->
            ( { model | size = Just size }
            , Process.sleep (250 * Time.millisecond)
                |> Task.perform (always <| ResizeTimer size)
                |> when (isJust model.size && isFresh model.graphState)
                |> Maybe.withDefault Cmd.none
            )

        ResizeTimer size ->
            ( model
            , case model.graphState of
                Fresh metric ->
                    if model.size == Just size then
                        sendMessage <| Render metric
                    else
                        Cmd.none

                _ ->
                    Cmd.none
            )

        DrawStarting metric ->
            ( { model | graphState = Drawing metric }, Cmd.none )

        DrawDone result ->
            ( if result.ok then
                { model | graphState = Fresh result.metric }
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

        Render metric ->
            let
                width =
                    model.size
                        |> Maybe.map .width
                        -- container padding is 15px :(
                        -- just have to keep this in sync because determining
                        -- dom sizes/offsets is really hard
                        |> Maybe.map (flip (-) 30)
            in
                ( model
                , Api.renderRequest metric
                    |> (\request -> { request | width = width })
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
        , Window.resizes Resize
        ]



-- VIEW


view : Model -> Html Msg
view model =
    Html.div []
        [ view_header
        , Html.div [ Attr.class "container" ]
            [ view_fullRow [ view_query model ]
            , view_fullRow [ view_canvas ]
            , view_fullRow [ view_errorText model ]
            ]
        ]


view_fullRow : List (Html msg) -> Html msg
view_fullRow children =
    Html.div [ Attr.class "row", Attr.style [ ( "padding-top", "10px" ) ] ]
        [ Html.div [ Attr.class "col-sm-12" ] children
        ]


view_centeredRow : List (Html msg) -> Html msg
view_centeredRow children =
    Html.div [ Attr.class "row", Attr.style [ ( "padding-top", "10px" ) ] ]
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



-- UTILS


when : Bool -> a -> Maybe a
when cond val =
    if cond then
        Just val
    else
        Nothing


isJust : Maybe a -> Bool
isJust val =
    case val of
        Nothing ->
            False

        Just _ ->
            True


sendMessage : msg -> Cmd msg
sendMessage msg =
    Task.perform (always msg) (Task.succeed ())
